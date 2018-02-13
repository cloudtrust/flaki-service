# Flaki service [![Build Status](https://travis-ci.org/cloudtrust/flaki-service.svg?branch=master)](https://travis-ci.org/cloudtrust/flaki-service) [![Coverage Status](https://coveralls.io/repos/github/cloudtrust/flaki-service/badge.svg?branch=master)](https://coveralls.io/github/cloudtrust/flaki-service?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/cloudtrust/flaki-service)](https://goreportcard.com/report/github.com/cloudtrust/flaki-service) [![OpenTracing Badge](https://img.shields.io/badge/OpenTracing-enabled-blue.svg)](http://opentracing.io)

Flaki service is a microservice that provides gRPC and HTTP access to [Flaki](https://github.com/cloudtrust/flaki), a distributed unique ID generator. 

The service includes logging, metrics, tracing, and error tracking. The logs are written to stdout and Redis in Logstash format for processing with the Elastic Stack.
Metrics such as number of IDs generated, time tracking,... are collected and saved to an InfluxDB Time Series Database.
Jaeger is used for distributed tracing and error tracking is managed with Sentry.

## Build
The service uses [FlatBuffers](https://google.github.io/flatbuffers/) for data serialisation. Make sure you have FlatBuffers installed and up to date with ```flatc --version```. It was tested with "flatc version 1.8.0 (Nov 22 2017)".

Build the service for the environment \<env>: 
```bash
./scripts/build.sh --env <env>
```
Note: \<env> is used for versioning. 

## Configuration
Configuration is done with a YAML file. An example is provided in ```./conf/DEV/flaki_service.yml```.

The following units can be configured: 
- component
- flaki
- redis
- influx
- sentry
- jaeger
- debug

Default configurations are provided, that is if an entry is not present in the configuration file, it will be set to its default value.

The following sections describe more precisely the available options.

### Component
For the component, the following parameters are available:

Key | Description | Default value 
--- | ----------- | ------------- 
component-name | name of the component | flaki-service 
component-http-address | HTTP server listening address | 0.0.0.0:8888 
component-grpc-address | gRPC server listening address  | 0.0.0.0:5555 

### Flaki
Key | Description | Default value 
--- | ----------- | ------------- 
flaki-node-id | node identifier | 0
flaki-component-id | component identidier | 0

If two Flaki instance have the same component ID and same node ID, there will be collisions on the generated IDs. So it is extremely important to initialise each instance of the Flaki generator with different node ID / component ID pairs, so we can ensure the uniqueness of the generated IDs.

More information on the Flaki unique ID generator are availaible on its [repository](https://github.com/cloudtrust/flaki).

### Redis
The logs can be formatted in [Logstash](https://www.elastic.co/products/logstash) format and send to a [Redis](https://redis.io/) server. The goal is to process them with the [Elastic Stack](https://www.elastic.co/products).

By default the Redis configuration is empty, which mean we do not use Redis. The loggers only log to stdout. 
If the Redis configuration keys are present in the configuration file, the logs are duplicated to stdout and Redis.

Key | Description | Default value 
--- | ----------- | ------------- 
redis-url | Redis server URL | ""
redis-password | Redis password | ""
redis-database | Redis database | 0

### InfluxDB
The service can record metrics and send them to an [Influx](https://www.influxdata.com/time-series-platform/influxdb/) time series DB.

By default the influx configuration is empty, which mean that the metrics are disabled. To activate them, provides configuration for infux in the configuration file.

Key | Description | Default value 
--- | ----------- | ------------- 
influx-url | Influx URL | ""
influx-username | InfluxDB username | ""
influx-password | InfluxDB password | ""
influx-database | InfluxDB database name | ""
influx-precision | Write precision of the points | ""
influx-retention-policy | Retention policy of the points | ""
influx-write-consistency | Number of servers required to confirm write | ""
influx-write-interval-ms | Flush interval in milliseconds | 1000

### Sentry
[Sentry](https://sentry.io/welcome/) is an open source error tracking system. It helps monitoring crashes and errors in real time.
By default, the error tracking is deactivated. To set it up, add the sentry-dsn to the configuration file.

Note: To obtain the sentry Data Source Name (DSN) you need to set up a project in Sentry (see [documenation](https://docs.sentry.io/quickstart/#configure-the-dsn)). 

Key | Description | Default value 
--- | ----------- | ------------- 
sentry-dsn | Sentry Data Source Name | ""

### Jaeger
[Jaeger](https://jaeger.readthedocs.io/en/latest/) is a distributed tracing system. It is disabled by default, to enable it provide configuration in the configuration file.

Key | Description | Default value 
--- | ----------- | ------------- 
jaeger-sampler-type |  | ""
jaeger-sampler-param |  | 0
jaeger-sampler-url |  | ""
jaeger-reporter-logspan |  | false
jaeger-write-interval-ms | Flush interval in milliseconds | 1000

### Debug
Key | Description | Default value 
--- | ----------- | ------------- 
pprof-route-enabled | whether the pprof debug routes are enabled | true

The golang pprof package serves runtime profiling data via the HTTP server (see [documentation](https://golang.org/pkg/net/http/pprof/)). If ```pprof-route-enabled``` is true, we enable the HTTP pprof routes.

## Usage
Launch the flaki service:
```bash
./bin/flaki_service --config-file <path/to/config/file.yml>
```
It is recommended to always provides an absolute path to the configuration file when the service is started, even though absolute and relative paths are supported.
If no configuration file is passed, the service will try to load the default config file at ```./conf/DEV/flaki_service.yml```, and if it fails it launches the service with the default parameters.

### gRPC

### HTTP

### Health
The service exposes HTTP routes to monitor the application health.
There is a root route returning the application general health, that is the list of components and whether they are "OK" or "KO".
Then each component has a dedicated route where more details are available: a set of tests and their results.

The root route is ```<component-http-address>/health``` and it returns the service general health as a JSON of the form:
```
{
  "influx": "OK",
  "redis": "OK",
  "sentry": "OK",
  "jaeger": "OK"
}
```

The subroutes are ```<component-http-address>/health/<name>``` and it returns the results of the tests for the component \<name>.
\<name> is the name of the component that matches the names in the JSON returned by the general route. In our case: "influx", "redis", "sentry", or "jaeger".
The subroutes return a JSON of the form:
```
[
  {
    "name": "ping",
    "duration": "906.881µs",
    "status": "OK"
  }
]
```
There is one entry per test, and each entry lists the name of the test, its duration and the status.

## About monitoring
Each gRPC or HTTP request will trigger a set of operations that are going to be logged, measured, tracked and traced. For those information to be usable, we must be able to link the logs, metrics, traces and error report together. We achieve that with a unique correlation ID. For a given request, the same correlation ID will appear on the logs, metrics, traces and error report.

Note: InfluxDB indexes tags, so we put the correlation ID as tags to speed up queries. To query a tag, do not forget to simple quote it, otherwise it always returns empty results.
```
select * from "<measurement>" where "correlation_id" = '<correlation_id>';
```

Note: In Jaeger UI, to search traces with a given correlation ID you must copy the following in the "Tags" box: 
```
correlation_id:<correlation_id>
```

## Tests

The unit tests don't cover:
- http client example (```./client/http/http.go```)
- grpc client example (```./client/grpc/grpc.go```)
- flakid  (```./cmd/flakid.go```)

The first two are provided as example.

The ```flakid.go``` is mosttly just the main function doing all the wiring, it is difficult to test it with unit tests. It is covered by our integration tests.

## Limitations

The Redis connection does not handle errors well: if there is a problem, it is closed forever. We will implement our own redis client later, because we need load-balancing and circuit-breaking.




