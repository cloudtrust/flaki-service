package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	component "github.com/JohanDroz/flaki-service/service/component"
	flaki_endpoint "github.com/JohanDroz/flaki-service/service/endpoint"
	module "github.com/JohanDroz/flaki-service/service/module"
	fb "github.com/JohanDroz/flaki-service/service/transport/flatbuffer/flaki"
	"github.com/JohanDroz/flaki-service/service/transport/grpc"
	flaki_http "github.com/JohanDroz/flaki-service/service/transport/http"
	"github.com/cloudtrust/flaki"
	sentry "github.com/getsentry/raven-go"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	gokit_influx "github.com/go-kit/kit/metrics/influx"
	"github.com/google/flatbuffers/go"
	"github.com/gorilla/mux"
	influx_client "github.com/influxdata/influxdb/client/v2"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	jaeger_client "github.com/uber/jaeger-client-go/config"
	"google.golang.org/grpc"
)

var (
	// Version of the component.
	Version = "1.0.0"
	// Environment is filled by the compiler.
	Environment = "unknown"
	// GitCommit is filled by the compiler.
	GitCommit = "unknown"
)

func main() {

	// Logger
	var logger = log.NewLogfmtLogger(os.Stdout)
	{
		logger = log.With(logger, "time", log.DefaultTimestampUTC, "caller", log.DefaultCaller)
		defer logger.Log("msg", "Goodbye")
	}

	// Configurations
	var config = config(log.With(logger, "component", "config_loader"))
	var (
		componentName    = fmt.Sprintf(config["component-name"].(string))
		grpcAddr         = fmt.Sprintf(config["component-grpc-address"].(string))
		httpAddr         = fmt.Sprintf(config["component-http-address"].(string))
		influxHTTPConfig = influx_client.HTTPConfig{
			Addr:     config["influx-url"].(string),
			Username: config["influx-username"].(string),
			Password: config["influx-password"].(string),
		}
		influxBatchPointsConfig = influx_client.BatchPointsConfig{
			Precision:        config["influx-precision"].(string),
			Database:         config["influx-database"].(string),
			RetentionPolicy:  config["influx-retention-policy"].(string),
			WriteConsistency: config["influx-write-consistency"].(string),
		}
		influxWriteInterval = time.Duration(config["influx-write-interval-ms"].(int)) * time.Millisecond
		jaegerConfig        = jaeger_client.Configuration{
			Sampler: &jaeger_client.SamplerConfig{
				Type:              config["jaeger-sampler-type"].(string),
				Param:             float64(config["jaeger-sampler-param"].(int)),
				SamplingServerURL: config["jaeger-sampler-url"].(string),
			},
			Reporter: &jaeger_client.ReporterConfig{
				LogSpans:            config["jaeger-reporter-logspan"].(bool),
				BufferFlushInterval: time.Duration(config["jaeger-reporter-flushinterval-ms"].(int)) * time.Millisecond,
			},
		}
		sentryDSN        = fmt.Sprintf(config["sentry-dsn"].(string))
		flakiNodeID      = uint64(config["flaki-node-id"].(int))
		flakiComponentID = uint64(config["flaki-component-id"].(int))
	)

	// Log component version infos
	logger.Log("component_name", componentName, "version", Version, "environment", Environment, "git_commit", GitCommit)

	// Critical errors channel
	var errc = make(chan error)
	go func() {
		var c = make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	// Flaki unique ID generator
	var flakiGen flaki.Flaki
	{
		var logger = log.With(logger, "component", "flaki")
		var err error
		flakiGen, err = flaki.NewFlaki(logger, flaki.ComponentID(flakiComponentID), flaki.NodeID(flakiNodeID))
		if err != nil {
			logger.Log("msg", "couldn't create flaki id generator", "error", err)
			return
		}
	}

	// Sentry
	var sentryClient *sentry.Client
	{
		var logger = log.With(logger, "component", "sentry")
		var err error
		logger.Log("sentry_dsn", sentryDSN)
		sentryClient, err = sentry.New(sentryDSN)
		if err != nil {
			logger.Log("msg", "Couldn't create Sentry client", "error", err)
			return
		}
		defer sentryClient.Close()
	}

	// Influx client
	var influxClient influx_client.Client
	{
		var logger = log.With(logger, "component", "influx")
		{
			var err error
			influxClient, err = influx_client.NewHTTPClient(influxHTTPConfig)
			if err != nil {
				logger.Log("msg", "Couldn't create Influx client", "error", err)
				return
			}
			defer influxClient.Close()
		}
	}

	// Influx go-kit handler
	var gokitInflux *gokit_influx.Influx
	{
		gokitInflux = gokit_influx.New(
			map[string]string{"service": "users"},
			influxBatchPointsConfig,
			log.With(logger, "component", "go-kit influx"),
		)
	}

	// Jaeger client
	var tracer opentracing.Tracer
	{
		var logger = log.With(logger, "component", "jaeger")
		var closer io.Closer
		var err error

		tracer, closer, err = jaegerConfig.New(componentName)
		if err != nil {
			logger.Log("error", err)
		}
		defer closer.Close()
	}
	opentracing.InitGlobalTracer(tracer)

	// Backend service
	var flakiModule module.Service
	{
		flakiModule = module.NewBasicService(flakiGen)
	}

	var flakiComponent component.Service
	{
		flakiComponent = component.NewBasicService(flakiModule)
		flakiComponent = component.MakeLoggingMiddleware(log.With(logger, "middleware", "component", "name", "flaki"))(flakiComponent)
	}

	var nextIDEndpoint endpoint.Endpoint
	{
		nextIDEndpoint = flaki_endpoint.MakeNextIDEndpoint(
			flakiComponent,
			flaki_endpoint.MakeMetricMiddleware(gokitInflux.NewHistogram("nextID-endpoint")),
			flaki_endpoint.MakeLoggingMiddleware(log.With(logger, "middleware", "endpoint", "method", "nextID")))
	}
	var nextValidIDEndpoint endpoint.Endpoint
	{
		nextValidIDEndpoint = flaki_endpoint.MakeNextValidIDEndpoint(
			flakiComponent,
			flaki_endpoint.MakeMetricMiddleware(gokitInflux.NewHistogram("nextValidID-endpoint")),
			flaki_endpoint.MakeLoggingMiddleware(log.With(logger, "middleware", "endpoint", "method", "nextValidID")))
	}

	var flakiEndpoints = flaki_endpoint.Endpoints{
		NextIDEndpoint:      nextIDEndpoint,
		NextValidIDEndpoint: nextValidIDEndpoint,
	}

	// GRPC server
	go func() {
		var logger = log.With(logger, "transport", "grpc")
		logger.Log("addr", grpcAddr)

		var flakiServer = grpc.NewServer(grpc.CustomCodec(flatbuffers.FlatbuffersCodec{}))
		var flakiGrpcServer = server.NewGrpcServer(flakiEndpoints)
		fb.RegisterFlakiServer(flakiServer, flakiGrpcServer)
		var lis net.Listener
		{
			var err error
			lis, err = net.Listen("tcp", grpcAddr)
			if err != nil {
				logger.Log("msg", "couldn't initialise listener", "error", err)
				errc <- err
				return
			}
		}
		errc <- flakiServer.Serve(lis)
	}()

	// HTTP server
	go func() {
		var logger = log.With(logger, "transport", "http")
		logger.Log("addr", httpAddr)

		var route = mux.NewRouter()

		// NextID
		route.Handle("/nextid", flaki_http.MakeNextIDHandler(flakiEndpoints.NextIDEndpoint, log.With(logger, "endpoint", "nextID")))

		// NextValidID
		route.Handle("/nextvalidid", flaki_http.MakeNextValidIDHandler(flakiEndpoints.NextValidIDEndpoint, log.With(logger, "endpoint", "nextValidID")))

		// Version
		route.Handle("/version", http.HandlerFunc(flaki_http.MakeVersion(componentName, Version, Environment, GitCommit)))

		errc <- http.ListenAndServe(httpAddr, route)
	}()

	// Influx writing
	go func() {
		var tic = time.NewTicker(influxWriteInterval)
		gokitInflux.WriteLoop(tic.C, influxClient)
	}()

	logger.Log("error", <-errc)
}

func config(logger log.Logger) map[string]interface{} {

	logger.Log("msg", "Loading configuration & command args")

	var configFile = "./conf/" + Environment + "/flaki_service.yml"
	// Component default
	viper.SetDefault("config-file", configFile)
	viper.SetDefault("component-name", "flaki-service")
	viper.SetDefault("component-http-address", "127.0.0.1:8888")
	viper.SetDefault("component-grpc-address", "127.0.0.1:5555")

	// Flaki generator default
	viper.SetDefault("flaki-node-id", 0)
	viper.SetDefault("flaki-component-id", 0)

	// Influx DB client default
	viper.SetDefault("influx-url", "http://localhost:8086")
	viper.SetDefault("influx-username", "admin")
	viper.SetDefault("influx-password", "admin")
	viper.SetDefault("influx-database", "metrics")
	viper.SetDefault("influx-precision", "s")
	viper.SetDefault("influx-retention-policy", "")
	viper.SetDefault("influx-write-consistency", "")
	viper.SetDefault("influx-write-interval-ms", 1000)

	// Sentry client default
	viper.SetDefault("sentry-dsn", "https://99360b38b8c947baaa222a5367cd74bc:579dc85095114b6198ab0f605d0dc576@sentry-cloudtrust.dev.elca.ch/2")

	// Jaeger tracing default
	viper.SetDefault("jaeger-sampler-type", "const")
	viper.SetDefault("jaeger-sampler-param", 1)
	viper.SetDefault("jaeger-sampler-url", "http://localhost:5775")
	viper.SetDefault("jaeger-reporter-logspan", false)
	viper.SetDefault("jaeger-reporter-flushinterval-ms", 1000)

	// First level of override
	pflag.String("config-file", viper.GetString("config-file"), "The configuration file path can be relative or absolute.")
	viper.BindPFlag("config-file", pflag.Lookup("config-file"))
	pflag.Parse()

	// Load & log config
	viper.SetConfigFile(viper.GetString("config-file"))
	var err = viper.ReadInConfig()
	if err != nil {
		logger.Log("error", err)
	}

	var config = viper.AllSettings()
	for k, v := range config {
		logger.Log(k, v)
	}

	return config
}
