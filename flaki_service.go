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

	"github.com/cloudtrust/flaki"
	flaki_component "github.com/cloudtrust/flaki-service/service/component"
	flaki_endpoint "github.com/cloudtrust/flaki-service/service/endpoint"
	flaki_module "github.com/cloudtrust/flaki-service/service/module"
	"github.com/cloudtrust/flaki-service/service/transport/flatbuffer/fb"
	flaki_grpc "github.com/cloudtrust/flaki-service/service/transport/grpc"
	flaki_http "github.com/cloudtrust/flaki-service/service/transport/http"
	"github.com/garyburd/redigo/redis"
	sentry "github.com/getsentry/raven-go"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	gokit_influx "github.com/go-kit/kit/metrics/influx"
	grpc_transport "github.com/go-kit/kit/transport/grpc"
	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/gorilla/mux"
	influx "github.com/influxdata/influxdb/client/v2"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	jaeger "github.com/uber/jaeger-client-go/config"
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

	// Logger.
	var logger = log.NewJSONLogger(os.Stdout)
	{
		logger = log.With(logger, "time", log.DefaultTimestampUTC, "caller", log.DefaultCaller)
	}

	// Configurations.
	var config = config(log.With(logger, "component", "config_loader"))
	var (
		componentName    = config["component-name"].(string)
		grpcAddr         = config["component-grpc-address"].(string)
		httpAddr         = config["component-http-address"].(string)
		influxHTTPConfig = influx.HTTPConfig{
			Addr:     fmt.Sprintf("http://%s", config["influx-url"].(string)),
			Username: config["influx-username"].(string),
			Password: config["influx-password"].(string),
		}
		influxBatchPointsConfig = influx.BatchPointsConfig{
			Precision:        config["influx-precision"].(string),
			Database:         config["influx-database"].(string),
			RetentionPolicy:  config["influx-retention-policy"].(string),
			WriteConsistency: config["influx-write-consistency"].(string),
		}
		influxWriteInterval = time.Duration(config["influx-write-interval-ms"].(int)) * time.Millisecond
		jaegerConfig        = jaeger.Configuration{
			Disabled: !config["jaeger"].(bool),
			Sampler: &jaeger.SamplerConfig{
				Type:              config["jaeger-sampler-type"].(string),
				Param:             float64(config["jaeger-sampler-param"].(int)),
				SamplingServerURL: fmt.Sprintf("http://%s", config["jaeger-sampler-url"].(string)),
			},
			Reporter: &jaeger.ReporterConfig{
				LogSpans:            config["jaeger-reporter-logspan"].(bool),
				BufferFlushInterval: time.Duration(config["jaeger-reporter-flushinterval-ms"].(int)) * time.Millisecond,
			},
		}
		sentryDSN        = fmt.Sprintf(config["sentry-dsn"].(string))
		flakiNodeID      = uint64(config["flaki-node-id"].(int))
		flakiComponentID = uint64(config["flaki-component-id"].(int))

		influxEnabled     = config["influx"].(bool)
		sentryEnabled     = config["sentry"].(bool)
		redisEnabled      = config["redis"].(bool)
		pprofRouteEnabled = config["pprof-route-enabled"].(bool)

		redisURL      = config["redis-url"].(string)
		redisPassword = config["redis-password"].(string)
		redisDatabase = config["redis-database"].(int)
	)

	// Health checks.
	var health = Health{}

	// Redis.
	if redisEnabled {
		var redisPool = &redis.Pool{
			Dial: func() (redis.Conn, error) {
				return redis.Dial("tcp", redisURL, redis.DialDatabase(redisDatabase), redis.DialPassword(redisPassword))
			},
		}
		defer redisPool.Close()

		// Create logger that duplicates logs to stdout and redis.
		logger = log.NewJSONLogger(io.MultiWriter(os.Stdout, NewLogstashRedisWriter(redisPool)))
		logger = log.With(logger, "time", log.DefaultTimestampUTC, "caller", log.DefaultCaller)
		defer logger.Log("msg", "Goodbye")

		// Redis health checks.
		health.AddCheck(MakeRedisHealthChecks(redisPool))
	}

	// Log component version infos.
	logger.Log("component_name", componentName, "version", Version, "environment", Environment, "git_commit", GitCommit)

	// Critical errors channel.
	var errc = make(chan error)
	go func() {
		var c = make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	// Flaki unique distributed ID generator.
	var flakiGen *flaki.Flaki
	{
		var logger = log.With(logger, "component", "flaki")
		var err error
		flakiGen, err = flaki.New(flaki.ComponentID(flakiComponentID), flaki.NodeID(flakiNodeID))
		if err != nil {
			logger.Log("msg", "couldn't create flaki id generator", "error", err)
			return
		}
	}

	// Sentry.
	type Sentry interface {
		CaptureError(err error, tags map[string]string, interfaces ...sentry.Interface) string
		CaptureErrorAndWait(err error, tags map[string]string, interfaces ...sentry.Interface) string
		URL() string
		Close()
	}

	var sentryClient Sentry
	if sentryEnabled {
		var logger = log.With(logger, "component", "sentry")
		var err error
		logger.Log("sentry_dsn", sentryDSN)
		sentryClient, err = sentry.New(sentryDSN)
		if err != nil {
			logger.Log("msg", "couldn't create Sentry client", "error", err)
			return
		}
		defer sentryClient.Close()
		health.AddCheck(MakeSentryHealthChecks(sentryClient))
	} else {
		sentryClient = &NoopSentry{}
	}

	// Influx client.
	type Metrics interface {
		NewCounter(name string) metrics.Counter
		NewGauge(name string) metrics.Gauge
		NewHistogram(name string) metrics.Histogram
		WriteLoop(c <-chan time.Time)
	}

	var influxMetrics Metrics
	if influxEnabled {
		var logger = log.With(logger, "component", "influx")

		var influxClient, err = influx.NewHTTPClient(influxHTTPConfig)
		if err != nil {
			logger.Log("msg", "couldn't create Influx client", "error", err)
			return
		}
		defer influxClient.Close()

		var gokitInflux = gokit_influx.New(
			map[string]string{},
			influxBatchPointsConfig,
			log.With(logger, "component", "go-kit influx"),
		)

		influxMetrics = NewMetrics(influxClient, gokitInflux)
		health.AddCheck(MakeInfluxHealthChecks(influxClient))
	} else {
		influxMetrics = &NoopMetrics{}
	}

	// Jaeger client.
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

	// Backend service.
	var flakiModule flaki_module.Service
	{
		flakiModule = flaki_module.NewBasicService(flakiGen)
		flakiModule = flaki_module.MakeLoggingMiddleware(log.With(logger, "middleware", "module"))(flakiModule)
		flakiModule = flaki_module.MakeTracingMiddleware(tracer)(flakiModule)
		flakiModule = flaki_module.MakeMetricMiddleware(influxMetrics.NewCounter("number_id"))(flakiModule)
	}

	var flakiComponent flaki_component.Service
	{
		flakiComponent = flaki_component.NewBasicService(flakiModule)
		flakiComponent = flaki_component.MakeLoggingMiddleware(log.With(logger, "middleware", "component"))(flakiComponent)
		flakiComponent = flaki_component.MakeErrorMiddleware(sentryClient)(flakiComponent)
		flakiComponent = flaki_component.MakeTracingMiddleware(tracer)(flakiComponent)
	}

	var flakiEndpoints = flaki_endpoint.NewEndpoints(flaki_endpoint.MakeCorrelationIDMiddleware(flakiGen))

	flakiEndpoints.MakeNextIDEndpoint(
		flakiComponent,
		flaki_endpoint.MakeMetricMiddleware(influxMetrics.NewHistogram("nextid_endpoint")),
		flaki_endpoint.MakeLoggingMiddleware(log.With(logger, "middleware", "endpoint", "method", "NextID")),
		flaki_endpoint.MakeTracingMiddleware(tracer, "nextid_endpoint"),
	)

	flakiEndpoints.MakeNextValidIDEndpoint(
		flakiComponent,
		flaki_endpoint.MakeMetricMiddleware(influxMetrics.NewHistogram("nextvalidid_endpoint")),
		flaki_endpoint.MakeLoggingMiddleware(log.With(logger, "middleware", "endpoint", "method", "NextValidID")),
		flaki_endpoint.MakeTracingMiddleware(tracer, "nextvalidid_endpoint"),
	)

	// GRPC server.
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

		// NextID.
		var nextIDHandler grpc_transport.Handler
		{
			nextIDHandler = flaki_grpc.MakeNextIDHandler(flakiEndpoints.NextIDEndpoint)
			nextIDHandler = flaki_grpc.MakeTracingMiddleware(tracer, "grpc_server_nextid")(nextIDHandler)
		}

		// NextValidID.
		var nextValidIDHandler grpc_transport.Handler
		{
			nextValidIDHandler = flaki_grpc.MakeNextValidIDHandler(flakiEndpoints.NextValidIDEndpoint)
			nextValidIDHandler = flaki_grpc.MakeTracingMiddleware(tracer, "grpc_server_nextvalidid")(nextValidIDHandler)
		}

		var grpcServer = flaki_grpc.NewGRPCServer(nextIDHandler, nextValidIDHandler)
		var flakiServer = grpc.NewServer(grpc.CustomCodec(flatbuffers.FlatbuffersCodec{}))
		fb.RegisterFlakiServer(flakiServer, grpcServer)

		errc <- flakiServer.Serve(lis)
	}()

	// HTTP server.
	go func() {
		var logger = log.With(logger, "transport", "http")
		logger.Log("addr", httpAddr)

		var route = mux.NewRouter()

		// NextID.
		var nextIDHandler http.Handler
		{
			nextIDHandler = flaki_http.MakeNextIDHandler(flakiEndpoints.NextIDEndpoint, tracer)
			nextIDHandler = flaki_http.MakeTracingMiddleware(tracer, "http_server_nextid")(nextIDHandler)
		}
		route.Handle("/nextid", nextIDHandler)

		// NextValidID.
		var nextValidIDHandler http.Handler
		{
			nextValidIDHandler = flaki_http.MakeNextValidIDHandler(flakiEndpoints.NextValidIDEndpoint, tracer)
			nextValidIDHandler = flaki_http.MakeTracingMiddleware(tracer, "http_server_nextvalidid")(nextValidIDHandler)
		}
		route.Handle("/nextvalidid", nextValidIDHandler)

		// Version.
		route.Handle("/", http.HandlerFunc(flaki_http.MakeVersion(componentName, Version, Environment, GitCommit)))

		// Health.
		health.RegisterRoutes(route)

		/*

			healthSubroute.HandleFunc("/", http.HandlerFunc(health.MakeHealthChecks(influxClient, sentryClient, tracer))
			healthSubroute.HandleFunc("/influx", http.HandlerFunc(flaki_http.MakeVersion("influx", "", "", "")))
			healthSubroute.HandleFunc("/sentry", http.HandlerFunc(flaki_http.MakeVersion("sentry", "", "", "")))
			healthSubroute.HandleFunc("/jaeger", http.HandlerFunc(flaki_http.MakeVersion("jaeger", "", "", "")))
		*/
		/* handle the following routes:
		/health/checks/: all checks, returns json like:
		{
			influx: up
			jaeger: up
			sentry: up
			...
		}
		then individual routes:
		/health/check/influx, sentry, jaeger, ....
		{
			nom du check: create db, ....
			temps: xxx ms
			status: OK/KO
		}
		*/

		// Debug.
		if pprofRouteEnabled {
			var debugSubroute = route.PathPrefix("/debug").Subrouter()
			debugSubroute.HandleFunc("/pprof/", http.HandlerFunc(pprof.Index))
			debugSubroute.HandleFunc("/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
			debugSubroute.HandleFunc("/pprof/profile", http.HandlerFunc(pprof.Profile))
			debugSubroute.HandleFunc("/pprof/symbol", http.HandlerFunc(pprof.Symbol))
			debugSubroute.HandleFunc("/pprof/trace", http.HandlerFunc(pprof.Trace))
		}

		errc <- http.ListenAndServe(httpAddr, route)
	}()

	// Influx writing.
	go func() {
		var tic = time.NewTicker(influxWriteInterval)
		influxMetrics.WriteLoop(tic.C)
	}()

	logger.Log("error", <-errc)
}

func config(logger log.Logger) map[string]interface{} {

	logger.Log("msg", "Loading configuration and command args")

	// Component default.
	viper.SetDefault("config-file", "./conf/DEV/flaki_service.yml")
	viper.SetDefault("component-name", "flaki-service")
	viper.SetDefault("component-http-address", "0.0.0.0:8888")
	viper.SetDefault("component-grpc-address", "0.0.0.0:5555")

	// Flaki generator default.
	viper.SetDefault("flaki-node-id", 0)
	viper.SetDefault("flaki-component-id", 0)

	// Influx DB client default.
	viper.SetDefault("influx", false)
	viper.SetDefault("influx-url", "")
	viper.SetDefault("influx-username", "")
	viper.SetDefault("influx-password", "")
	viper.SetDefault("influx-database", "")
	viper.SetDefault("influx-precision", "")
	viper.SetDefault("influx-retention-policy", "")
	viper.SetDefault("influx-write-consistency", "")
	viper.SetDefault("influx-write-interval-ms", 1000)

	// Sentry client default.
	viper.SetDefault("sentry", false)
	viper.SetDefault("sentry-dsn", "")

	// Jaeger tracing default.
	viper.SetDefault("jaeger", false)
	viper.SetDefault("jaeger-sampler-type", "")
	viper.SetDefault("jaeger-sampler-param", 0)
	viper.SetDefault("jaeger-sampler-url", "")
	viper.SetDefault("jaeger-reporter-logspan", false)
	viper.SetDefault("jaeger-reporter-flushinterval-ms", 1000)

	// Debug routes enabled.
	viper.SetDefault("pprof-route-enabled", true)

	// Redis.
	viper.SetDefault("redis", false)
	viper.SetDefault("redis-url", "")
	viper.SetDefault("redis-password", "")
	viper.SetDefault("redis-database", 0)

	// First level of override.
	pflag.String("config-file", viper.GetString("config-file"), "The configuration file path can be relative or absolute.")
	viper.BindPFlag("config-file", pflag.Lookup("config-file"))
	pflag.Parse()

	// Load config.
	viper.SetConfigFile(viper.GetString("config-file"))
	var err = viper.ReadInConfig()
	if err != nil {
		logger.Log("error", err)
	}
	var config = viper.AllSettings()

	// If the URL is not set, we consider the components disabled.
	config["influx"] = config["influx-url"].(string) != ""
	config["sentry"] = config["sentry-dsn"].(string) != ""
	config["jaeger"] = config["jaeger-sampler-url"].(string) != ""
	config["redis"] = config["redis-url"].(string) != ""

	for k, v := range config {
		logger.Log(k, v)
	}

	return config
}
