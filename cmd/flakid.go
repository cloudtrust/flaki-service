package main

import (
	"encoding/json"
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
	"github.com/cloudtrust/flaki-service/pkg/flaki"
	"github.com/cloudtrust/flaki-service/pkg/flaki/flatbuffer/fb"
	"github.com/cloudtrust/flaki-service/pkg/health"
	"github.com/cloudtrust/flaki-service/pkg/middleware"
	"github.com/garyburd/redigo/redis"
	sentry "github.com/getsentry/raven-go"
	"github.com/go-kit/kit/endpoint"
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
		logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)
	}

	// Configurations.
	var config = config(log.With(logger, "unit", "config"))
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
				BufferFlushInterval: time.Duration(config["jaeger-write-interval-ms"].(int)) * time.Millisecond,
			},
		}
		sentryDSN        = fmt.Sprintf(config["sentry-dsn"].(string))
		flakiNodeID      = uint64(config["flaki-node-id"].(int))
		flakiComponentID = uint64(config["flaki-component-id"].(int))

		influxEnabled     = config["influx"].(bool)
		sentryEnabled     = config["sentry"].(bool)
		redisEnabled      = config["redis"].(bool)
		pprofRouteEnabled = config["pprof-route-enabled"].(bool)

		redisURL           = config["redis-url"].(string)
		redisPassword      = config["redis-password"].(string)
		redisDatabase      = config["redis-database"].(int)
		redisWriteInterval = time.Duration(config["redis-write-interval-ms"].(int)) * time.Millisecond
	)

	// Redis.
	var redisConn redis.Conn
	if redisEnabled {
		var err error
		redisConn, err = redis.Dial("tcp", redisURL, redis.DialDatabase(redisDatabase), redis.DialPassword(redisPassword))
		if err != nil {
			logger.Log("msg", "could not create redis client", "error", err)
			return
		}
		defer redisConn.Close()

		// Create logger that duplicates logs to stdout and redis.
		logger = log.NewJSONLogger(io.MultiWriter(os.Stdout, NewLogstashRedisWriter(redisConn)))
		logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)
	}
	defer logger.Log("msg", "goodbye")

	// Add component name and version to the logger tags.
	logger = log.With(logger, "component_name", componentName, "component_version", Version)

	// Log component version infos.
	logger.Log("environment", Environment, "git_commit", GitCommit)

	// Critical errors channel.
	var errc = make(chan error)
	go func() {
		var c = make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	// Flaki unique distributed ID generator.
	var flakiGen *flaki_gen.Flaki
	{
		var logger = log.With(logger, "unit", "flaki")
		var err error
		flakiGen, err = flaki_gen.New(flaki_gen.ComponentID(flakiComponentID), flaki_gen.NodeID(flakiNodeID))
		if err != nil {
			logger.Log("msg", "could not create Flaki generator", "error", err)
			return
		}
	}

	// Sentry.
	type Sentry interface {
		CaptureError(err error, tags map[string]string, interfaces ...sentry.Interface) string
		URL() string
		Close()
	}

	var sentryClient Sentry
	if sentryEnabled {
		var logger = log.With(logger, "unit", "sentry")
		var err error
		sentryClient, err = sentry.New(sentryDSN)
		if err != nil {
			logger.Log("msg", "could not create Sentry client", "error", err)
			return
		}
		defer sentryClient.Close()
	} else {
		sentryClient = &NoopSentry{}
	}

	// Influx client.
	type Metrics interface {
		NewCounter(name string) metrics.Counter
		NewGauge(name string) metrics.Gauge
		NewHistogram(name string) metrics.Histogram
		WriteLoop(c <-chan time.Time)
		Ping(timeout time.Duration) (time.Duration, string, error)
	}

	var influxMetrics Metrics
	if influxEnabled {
		var logger = log.With(logger, "unit", "influx")

		var influxClient, err = influx.NewHTTPClient(influxHTTPConfig)
		if err != nil {
			logger.Log("msg", "could not create Influx client", "error", err)
			return
		}
		defer influxClient.Close()

		var gokitInflux = gokit_influx.New(
			map[string]string{},
			influxBatchPointsConfig,
			log.With(logger, "unit", "go-kit influx"),
		)

		influxMetrics = NewMetrics(influxClient, gokitInflux)
	} else {
		influxMetrics = &NoopMetrics{}
	}

	// Jaeger client.
	var tracer opentracing.Tracer
	{
		var logger = log.With(logger, "unit", "jaeger")
		var closer io.Closer
		var err error

		tracer, closer, err = jaegerConfig.New(componentName)
		if err != nil {
			logger.Log("msg", "could not create Jaeger tracer", "error", err)
			return
		}
		defer closer.Close()

	}

	// Flaki service.
	var flakiLogger = log.With(logger, "svc", "flaki")

	var flakiModule flaki.Module
	{
		flakiModule = flaki.NewModule(flakiGen)
		flakiModule = flaki.MakeModuleInstrumentingCounterMW(influxMetrics.NewCounter("flaki_module_ctr"))(flakiModule)
		flakiModule = flaki.MakeModuleInstrumentingMW(influxMetrics.NewHistogram("flaki_module"))(flakiModule)
		flakiModule = flaki.MakeModuleLoggingMW(log.With(flakiLogger, "mw", "module"))(flakiModule)
		flakiModule = flaki.MakeModuleTracingMW(tracer)(flakiModule)
	}

	var flakiComponent flaki.Component
	{
		flakiComponent = flaki.NewComponent(flakiModule)
		flakiComponent = flaki.MakeComponentInstrumentingMW(influxMetrics.NewHistogram("flaki_component"))(flakiComponent)
		flakiComponent = flaki.MakeComponentLoggingMW(log.With(flakiLogger, "mw", "component"))(flakiComponent)
		flakiComponent = flaki.MakeComponentTracingMW(tracer)(flakiComponent)
		flakiComponent = flaki.MakeComponentTrackingMW(sentryClient)(flakiComponent)
	}

	var nextIDEndpoint endpoint.Endpoint
	{
		nextIDEndpoint = flaki.MakeNextIDEndpoint(flakiComponent)
		nextIDEndpoint = middleware.MakeEndpointInstrumentingMW(influxMetrics.NewHistogram("nextid_endpoint"))(nextIDEndpoint)
		nextIDEndpoint = middleware.MakeEndpointLoggingMW(log.With(flakiLogger, "mw", "endpoint", "unit", "NextID"))(nextIDEndpoint)
		nextIDEndpoint = middleware.MakeEndpointTracingMW(tracer, "nextid_endpoint")(nextIDEndpoint)
	}

	var nextValidIDEndpoint endpoint.Endpoint
	{
		nextValidIDEndpoint = flaki.MakeNextValidIDEndpoint(flakiComponent)
		nextValidIDEndpoint = middleware.MakeEndpointInstrumentingMW(influxMetrics.NewHistogram("nextvalidid_endpoint"))(nextValidIDEndpoint)
		nextValidIDEndpoint = middleware.MakeEndpointLoggingMW(log.With(flakiLogger, "mw", "endpoint", "unit", "NextValidID"))(nextValidIDEndpoint)
		nextValidIDEndpoint = middleware.MakeEndpointTracingMW(tracer, "nextvalidid_endpoint")(nextValidIDEndpoint)
	}

	var flakiEndpoints = flaki.Endpoints{
		NextIDEndpoint:      nextIDEndpoint,
		NextValidIDEndpoint: nextValidIDEndpoint,
	}

	// Health service.
	var healthLogger = log.With(logger, "svc", "health")

	var healthComponent health.Component
	{
		var influxHM = health.NewInfluxModule(influxMetrics)
		var jaegerHM = health.NewJaegerModule(tracer)
		var redisHM = health.NewRedisModule(redisConn)
		var sentryHM = health.NewSentryModule(sentryClient)

		healthComponent = health.NewComponent(influxHM, jaegerHM, redisHM, sentryHM)
	}

	var influxHealthEndpoint endpoint.Endpoint
	{
		influxHealthEndpoint = health.MakeInfluxHealthCheckEndpoint(healthComponent)
		influxHealthEndpoint = middleware.MakeEndpointInstrumentingMW(influxMetrics.NewHistogram("influx_health_endpoint"))(influxHealthEndpoint)
		influxHealthEndpoint = middleware.MakeEndpointLoggingMW(log.With(healthLogger, "mw", "endpoint", "unit", "influx"))(influxHealthEndpoint)
		influxHealthEndpoint = middleware.MakeEndpointTracingMW(tracer, "influx_health_endpoint")(influxHealthEndpoint)
		influxHealthEndpoint = middleware.MakeEndpointCorrelationIDMW(flakiEndpoints)(influxHealthEndpoint)
	}
	var jaegerHealthEndpoint endpoint.Endpoint
	{
		jaegerHealthEndpoint = health.MakeJaegerHealthCheckEndpoint(healthComponent)
		jaegerHealthEndpoint = middleware.MakeEndpointInstrumentingMW(influxMetrics.NewHistogram("jaeger_health_endpoint"))(jaegerHealthEndpoint)
		jaegerHealthEndpoint = middleware.MakeEndpointLoggingMW(log.With(healthLogger, "mw", "endpoint", "unit", "jaeger"))(jaegerHealthEndpoint)
		jaegerHealthEndpoint = middleware.MakeEndpointTracingMW(tracer, "jaeger_health_endpoint")(jaegerHealthEndpoint)
		jaegerHealthEndpoint = middleware.MakeEndpointCorrelationIDMW(flakiEndpoints)(jaegerHealthEndpoint)
	}
	var redisHealthEndpoint endpoint.Endpoint
	{
		redisHealthEndpoint = health.MakeRedisHealthCheckEndpoint(healthComponent)
		redisHealthEndpoint = middleware.MakeEndpointInstrumentingMW(influxMetrics.NewHistogram("redis_health_endpoint"))(redisHealthEndpoint)
		redisHealthEndpoint = middleware.MakeEndpointLoggingMW(log.With(healthLogger, "mw", "endpoint", "unit", "redis"))(redisHealthEndpoint)
		redisHealthEndpoint = middleware.MakeEndpointTracingMW(tracer, "redis_health_endpoint")(redisHealthEndpoint)
		redisHealthEndpoint = middleware.MakeEndpointCorrelationIDMW(flakiEndpoints)(redisHealthEndpoint)
	}
	var sentryHealthEndpoint endpoint.Endpoint
	{
		sentryHealthEndpoint = health.MakeSentryHealthCheckEndpoint(healthComponent)
		sentryHealthEndpoint = middleware.MakeEndpointInstrumentingMW(influxMetrics.NewHistogram("sentry_health_endpoint"))(sentryHealthEndpoint)
		sentryHealthEndpoint = middleware.MakeEndpointLoggingMW(log.With(healthLogger, "mw", "endpoint", "unit", "sentry"))(sentryHealthEndpoint)
		sentryHealthEndpoint = middleware.MakeEndpointTracingMW(tracer, "sentry_health_endpoint")(sentryHealthEndpoint)
		sentryHealthEndpoint = middleware.MakeEndpointCorrelationIDMW(flakiEndpoints)(sentryHealthEndpoint)
	}

	var healthEndpoints = health.Endpoints{
		InfluxHealthCheck: influxHealthEndpoint,
		JaegerHealthCheck: jaegerHealthEndpoint,
		RedisHealthCheck:  redisHealthEndpoint,
		SentryHealthCheck: sentryHealthEndpoint,
	}

	// GRPC server.
	go func() {
		var logger = log.With(logger, "transport", "grpc")
		logger.Log("addr", grpcAddr)

		var lis net.Listener
		{
			var err error
			lis, err = net.Listen("tcp", grpcAddr)
			if err != nil {
				logger.Log("msg", "could not initialise listener", "error", err)
				errc <- err
				return
			}
		}

		// NextID.
		var nextIDHandler grpc_transport.Handler
		{
			nextIDHandler = flaki.MakeGRPCNextIDHandler(flakiEndpoints.NextIDEndpoint)
			nextIDHandler = flaki.MakeGRPCTracingMW(tracer, componentName, "grpc_server_nextid")(nextIDHandler)
		}

		// NextValidID.
		var nextValidIDHandler grpc_transport.Handler
		{
			nextValidIDHandler = flaki.MakeGRPCNextValidIDHandler(flakiEndpoints.NextValidIDEndpoint)
			nextValidIDHandler = flaki.MakeGRPCTracingMW(tracer, componentName, "grpc_server_nextvalidid")(nextValidIDHandler)
		}

		var grpcServer = flaki.NewGRPCServer(nextIDHandler, nextValidIDHandler)
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
			nextIDHandler = flaki.MakeHTTPNextIDHandler(flakiEndpoints.NextIDEndpoint)
			nextIDHandler = flaki.MakeHTTPTracingMW(tracer, componentName, "http_server_nextid")(nextIDHandler)
		}
		route.Handle("/nextid", nextIDHandler)

		// NextValidID.
		var nextValidIDHandler http.Handler
		{
			nextValidIDHandler = flaki.MakeHTTPNextValidIDHandler(flakiEndpoints.NextValidIDEndpoint)
			nextValidIDHandler = flaki.MakeHTTPTracingMW(tracer, componentName, "http_server_nextvalidid")(nextValidIDHandler)
		}
		route.Handle("/nextvalidid", nextValidIDHandler)

		// Version.
		route.Handle("/", http.HandlerFunc(makeVersion(componentName, Version, Environment, GitCommit)))

		// Health checks.
		var healthSubroute = route.PathPrefix("/health").Subrouter()

		healthSubroute.Handle("", http.HandlerFunc(health.MakeHealthChecksHandler(healthEndpoints)))

		var influxHealthCheckHandler = health.MakeInfluxHealthCheckHandler(healthEndpoints.InfluxHealthCheck)
		healthSubroute.Handle("/influx", influxHealthCheckHandler)

		var jaegerHealthCheckHandler = health.MakeJaegerHealthCheckHandler(healthEndpoints.JaegerHealthCheck)
		healthSubroute.Handle("/jaeger", jaegerHealthCheckHandler)

		var redisHealthCheckHandler = health.MakeRedisHealthCheckHandler(healthEndpoints.RedisHealthCheck)
		healthSubroute.Handle("/redis", redisHealthCheckHandler)

		var sentryHealthCheckHandler = health.MakeSentryHealthCheckHandler(healthEndpoints.SentryHealthCheck)
		healthSubroute.Handle("/sentry", sentryHealthCheckHandler)

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
		defer tic.Stop()
		influxMetrics.WriteLoop(tic.C)
	}()

	// Redis writing.
	if redisEnabled {
		go func() {
			var tic = time.NewTicker(redisWriteInterval)
			defer tic.Stop()
			for range tic.C {
				redisConn.Flush()
			}
		}()
	}
	logger.Log("error", <-errc)
}

type info struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Env     string `json:"environment"`
	Commit  string `json:"commit"`
}

// makeVersion makes a HTTP handler that returns information about the version of the service.
func makeVersion(componentName, version, environment, gitCommit string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		var infos = info{
			Name:    componentName,
			Version: version,
			Env:     environment,
			Commit:  gitCommit,
		}

		var j, err = json.MarshalIndent(infos, "", "  ")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write(j)
		}
	}
}

func config(logger log.Logger) map[string]interface{} {
	logger.Log("msg", "load configuration and command args")

	// Component default.
	viper.SetDefault("config-file", "./conf/DEV/flakid.yml")
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
	viper.SetDefault("jaeger-write-interval-ms", 1000)

	// Debug routes enabled.
	viper.SetDefault("pprof-route-enabled", true)

	// Redis.
	viper.SetDefault("redis", false)
	viper.SetDefault("redis-url", "")
	viper.SetDefault("redis-password", "")
	viper.SetDefault("redis-database", 0)
	viper.SetDefault("redis-write-interval-ms", 1000)

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
