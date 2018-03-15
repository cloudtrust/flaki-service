# Flaki service [![Build Status][ci-img]][ci] [![Coverage Status][cov-img]][cov] [![GoDoc][godoc-img]][godoc] [![Go Report Card][report-img]][report] [![OpenTracing Badge][opentracing-img]][opentracing]

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

A dockerfile is provided to make the flaki-service run in a container.
The container contains the flaki-service itself, the jaeger agent that handle traces and monit.
To build the container you must provides the following build-args:

- `flaki_service_git_tag`: the git tag of flaki-service repository (<https://github.com/cloudtrust/flaki-service>).
- `flaki_service_release`: the flaki-service release archive, e.g. <https://github.com/cloudtrust/flaki-service/releases/download/1.0/v1.0.tar.gz>. It can be found [here](https://github.com/cloudtrust/flaki-service/releases).
- `jaeger_release`: the jaeger release archive, e.g. <https://github.com/cloudtrust/jaeger/releases/download/v1.2.0/v1.2.0.tar.gz>. It can be found [here](https://github.com/cloudtrust/jaeger/releases).
- `config_repo`: the repository containing the configuration, e.g. <https://github.com/cloudtrust/dev-config.git>.
- `config_git_tag`: the git tag of config repository.

Then you can build the image.

```bash
mkdir build_context
cp dockerfiles/cloudtrust-flaki.dockerfile build_context/
cd build_context

docker build --build-arg flaki_service_git_tag=<flaki_service_git_tag> --build-arg flaki_service_release=<flaki_service_release> --build-arg jaeger_release=<jaeger_release> --build-arg config_git_tag=<config_git_tag> --build-arg config_repo=<config_repo> -t cloudtrust-flaki-service -f cloudtrust-flaki.dockerfile .
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
component-http-host-port | HTTP server listening address | 0.0.0.0:8888
component-grpc-host-port | gRPC server listening address  | 0.0.0.0:5555

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
./bin/flakid --config-file <path/to/config/file.yml>
```

It is recommended to always provides an absolute path to the configuration file when the service is started, even though absolute and relative paths are supported.
If no configuration file is passed, the service will try to load the default config file at ```./conf/DEV/flakid.yml```, and if it fails it launches the service with the default parameters.

### gRPC and HTTP clients

To obtain IDs using gRPC or HTTP, you need to implement your own clients. There is an example in the directory `client`.
There are two methods available to get IDs: NextID and NextValidID. Both take a Flatbuffer `FlakiRequest` and reply with a Flatbuffer `FlakiReply` containing the unique ID. The Flatbuffer schema is `pkg/flaki/flatbuffer/flaki.fbs`.

### Health

The service exposes HTTP routes to monitor the application health.
There is a root route returning the application general health, that is the list of components and whether they are "OK" or "KO".
Then each component has a dedicated route where more details are available: a set of tests and their results.

The root route is ```<component-http-host-port>/health``` and it returns the service general health as a JSON of the form:

```json
{
  "influx": "OK",
  "redis": "OK",
  "sentry": "OK",
  "jaeger": "OK"
}
```

The subroutes are ```<component-http-host-port>/health/<name>``` and it returns the results of the tests for the component \<name>.
\<name> is the name of the component that matches the names in the JSON returned by the general route. In our case: "influx", "redis", "sentry", or "jaeger".
The subroutes return a JSON of the form:

```json
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

```sql
select * from "<measurement>" where "correlation_id" = '<correlation_id>';
```

Note: In Jaeger UI, to search traces with a given correlation ID you must copy the following in the "Tags" box:

```sql
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

[ci-img]: https://travis-ci.org/cloudtrust/flaki-service.svg?branch=master
[ci]: https://travis-ci.org/cloudtrust/flaki-service
[cov-img]: https://coveralls.io/repos/github/cloudtrust/flaki-service/badge.svg?branch=master
[cov]: https://coveralls.io/github/cloudtrust/flaki-service?branch=master
[godoc-img]: https://godoc.org/github.com/cloudtrust/flaki-service?status.svg
[godoc]: https://godoc.org/github.com/cloudtrust/flaki-service
[report-img]: https://goreportcard.com/badge/github.com/cloudtrust/flaki-service
[report]: https://goreportcard.com/report/github.com/cloudtrust/flaki-service
[opentracing-img]: https://img.shields.io/badge/OpenTracing-enabled-blue.svg
[opentracing]: http://opentracing.io
