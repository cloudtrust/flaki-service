package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	flaki_gen "github.com/cloudtrust/flaki"
	component "github.com/cloudtrust/flaki-service/service/component"
	flaki_endpoint "github.com/cloudtrust/flaki-service/service/endpoint"
	module "github.com/cloudtrust/flaki-service/service/module"
	fb "github.com/cloudtrust/flaki-service/service/transport/flatbuffer/flaki"
	flaki_grpc "github.com/cloudtrust/flaki-service/service/transport/grpc"
	flaki_http "github.com/cloudtrust/flaki-service/service/transport/http"
	sentry "github.com/getsentry/raven-go"
	"github.com/go-kit/kit/log"
	gokit_influx "github.com/go-kit/kit/metrics/influx"
	grpc_transport "github.com/go-kit/kit/transport/grpc"
	flatbuffers "github.com/google/flatbuffers/go"
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

	// Flaki unique distributed ID generator
	var flaki flaki_gen.Flaki
	{
		var logger = log.With(logger, "component", "flaki")
		var err error
		flaki, err = flaki_gen.NewFlaki(flaki_gen.ComponentID(flakiComponentID), flaki_gen.NodeID(flakiNodeID))
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
			map[string]string{},
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
			return
		}
		defer closer.Close()
	}

	// Backend service
	var flakiModule module.Service
	{
		flakiModule = module.NewBasicService(flaki)
	}

	var flakiComponent component.Service
	{
		flakiComponent = component.NewBasicService(flakiModule)
		flakiComponent = component.MakeLoggingMiddleware(log.With(logger, "middleware", "component", "name", "flaki"))(flakiComponent)
		flakiComponent = component.MakeErrorMiddleware(sentryClient)(flakiComponent)
	}

	var flakiEndpoints = flaki_endpoint.NewEndpoints(flaki_endpoint.MakeCorrelationIDMiddleware(flaki))

	flakiEndpoints.MakeNextIDEndpoint(
		flakiComponent,
		flaki_endpoint.MakeMetricMiddleware(gokitInflux.NewHistogram("nextID-endpoint")),
		flaki_endpoint.MakeLoggingMiddleware(log.With(logger, "middleware", "endpoint", "method", "nextID")),
		flaki_endpoint.MakeTracingMiddleware(tracer, "nextID"),
	)

	flakiEndpoints.MakeNextValidIDEndpoint(
		flakiComponent,
		flaki_endpoint.MakeMetricMiddleware(gokitInflux.NewHistogram("nextValidID-endpoint")),
		flaki_endpoint.MakeLoggingMiddleware(log.With(logger, "middleware", "endpoint", "method", "nextValidID")),
		flaki_endpoint.MakeTracingMiddleware(tracer, "nextValidID"),
	)

	// GRPC server
	go func() {
		var logger = log.With(logger, "transport", "grpc")
		logger.Log("addr", grpcAddr)

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

		// NextID
		var nextIDHandler grpc_transport.Handler
		{
			var logger = log.With(logger, "endpoint", "nextID")
			nextIDHandler = flaki_grpc.MakeNextIDHandler(flakiEndpoints.NextIDEndpoint, logger, tracer)
		}

		// NextValidID
		var nextValidIDHandler grpc_transport.Handler
		{
			var logger = log.With(logger, "endpoint", "nextValidID")
			nextValidIDHandler = flaki_grpc.MakeNextValidIDHandler(flakiEndpoints.NextValidIDEndpoint, logger, tracer)
		}

		var grpcServer = flaki_grpc.NewGRPCServer(nextIDHandler, nextValidIDHandler)
		var flakiServer = grpc.NewServer(grpc.CustomCodec(flatbuffers.FlatbuffersCodec{}))
		fb.RegisterFlakiServer(flakiServer, grpcServer)

		errc <- flakiServer.Serve(lis)
	}()

	// HTTP server
	go func() {
		var logger = log.With(logger, "transport", "http")
		logger.Log("addr", httpAddr)

		var route = mux.NewRouter()

		// NextID
		var nextIDHandler http.Handler
		{
			var logger = log.With(logger, "endpoint", "nextID")
			nextIDHandler = flaki_http.MakeNextIDHandler(flakiEndpoints.NextIDEndpoint, logger, tracer)
		}
		route.Handle("/nextid", nextIDHandler)

		// NextValidID
		var nextValidIDHandler http.Handler
		{
			var logger = log.With(logger, "endpoint", "nextValidID")
			nextValidIDHandler = flaki_http.MakeNextValidIDHandler(flakiEndpoints.NextValidIDEndpoint, logger, tracer)
		}
		route.Handle("/nextvalidid", nextValidIDHandler)

		// Version
		route.Handle("/version", http.HandlerFunc(flaki_http.MakeVersion(componentName, Version, Environment, GitCommit)))

		// Debug
		var debugSubroute = route.PathPrefix("/debug").Subrouter()
		debugSubroute.HandleFunc("/pprof/", http.HandlerFunc(pprof.Index))
		debugSubroute.HandleFunc("/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
		debugSubroute.HandleFunc("/pprof/profile", http.HandlerFunc(pprof.Profile))
		debugSubroute.HandleFunc("/pprof/symbol", http.HandlerFunc(pprof.Symbol))
		debugSubroute.HandleFunc("/pprof/trace", http.HandlerFunc(pprof.Trace))

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

	var configFile = "./conf/DEV/flaki_service.yml"
	// Component default
	viper.SetDefault("config-file", configFile)
	viper.SetDefault("component-name", "flaki-service")
	viper.SetDefault("component-http-address", "0.0.0.0:8888")
	viper.SetDefault("component-grpc-address", "0.0.0.0:5555")

	// Flaki generator default
	viper.SetDefault("flaki-node-id", 0)
	viper.SetDefault("flaki-component-id", 0)

	// Influx DB client default
	viper.SetDefault("influx-url", "http://127.0.0.1:8086")
	viper.SetDefault("influx-username", "flaki")
	viper.SetDefault("influx-password", "flaki")
	viper.SetDefault("influx-database", "flakimetrics")
	viper.SetDefault("influx-precision", "s")
	viper.SetDefault("influx-retention-policy", "")
	viper.SetDefault("influx-write-consistency", "")
	viper.SetDefault("influx-write-interval-ms", 1000)

	// Sentry client default
	viper.SetDefault("sentry-dsn", "")

	// Jaeger tracing default
	viper.SetDefault("jaeger-sampler-type", "const")
	viper.SetDefault("jaeger-sampler-param", 1)
	viper.SetDefault("jaeger-sampler-url", "http://127.0.0.1:5775")
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
