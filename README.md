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

## Container
If you want to run the flaki-service in a docker container:
```bash
mkdir build_context
cp bin/flakid build_context/
cp dockerfiles/cloudtrust-flaki.dockerfile build_context/
cd build_context

# For the tracing, you also need to put the jaeger "agent-linux" executable in the build_context.

#Build the dockerfile for DEV environment
docker build --build-arg flaki_service_git_tag=<git_tag> -t cloudtrust-flaki-service -f cloudtrust-flaki.dockerfile .
docker create --tmpfs /tmp --tmpfs /run -v /sys/fs/cgroup:/sys/fs/cgroup:ro -p 5555:5555 -p 8888:8888 --name flaki-service-1 cloudtrust-flaki-service
```

## Configuration
Configuration is done with a YAML file, e.g. ```./conf/DEV/flakid.yml```.
Default configurations are provided, that is if an entry is not present in the configuration file, it will be set to its default value.

The documentation for the [Redis](https://cloudtrust.github.io/doc/chapter-godevel/logging.html), [Influx](https://cloudtrust.github.io/doc/chapter-godevel/instrumenting.html), [Sentry](https://cloudtrust.github.io/doc/chapter-godevel/tracking.html), [Jaeger](https://cloudtrust.github.io/doc/chapter-godevel/tracing.html) and [Debug](https://cloudtrust.github.io/doc/chapter-godevel/debugging.html) configuration are common to all microservices and is provided in the Cloudtrust Gitbook.

The configurations specific to the flaki-service are described in the next sections.

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

## Usage
Launch the flaki service:
```bash
./bin/flaki_service --config-file <path/to/config/file.yml>
```
It is recommended to always provides an absolute path to the configuration file when the service is started, even though absolute and relative paths are supported.
If no configuration file is passed, the service will try to load the default config file at ```./conf/DEV/flaki_service.yml```, and if it fails it launches the service with the default parameters.

### gRPC and HTTP clients
To obtain IDs using gRPC or HTTP, you need to implement your own clients. There is an example in the directory `client`.
There are two methods available to get IDs: NextID and NextValidID. Both take a Flatbuffer `EmptyRequest` and reply with a Flatbuffer `FlakiReply` containing the unique ID and an error. The Flatbuffer schema is `pkg/flaki/flatbuffer/flaki.fbs`.

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

Gomock is used to automatically genarate mocks. See the Cloudtrust [Gitbook](https://cloudtrust.github.io/doc/chapter-godevel/testing.html) for more information.

The unit tests don't cover:
- http client example (```./client/http/http.go```)
- grpc client example (```./client/grpc/grpc.go```)
- flakid  (```./cmd/flakid.go```)

The first two are provided as example.

The ```flakid.go``` is mostly just the main function doing all the wiring, it is difficult to test it with unit tests. It is covered by our integration tests.

## Limitations

The Redis connection does not handle errors well: if there is a problem, it is closed forever. We will implement our own redis client later, because we need load-balancing and circuit-breaking.
