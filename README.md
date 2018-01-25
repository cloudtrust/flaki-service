# Flaki service [![Build Status](https://travis-ci.org/cloudtrust/flaki-service.svg?branch=master)](https://travis-ci.org/cloudtrust/flaki-service) [![Coverage Status](https://coveralls.io/repos/github/cloudtrust/flaki-service/badge.svg?branch=master)](https://coveralls.io/github/cloudtrust/flaki-service?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/cloudtrust/flaki-service)](https://goreportcard.com/report/github.com/cloudtrust/flaki-service) [![OpenTracing Badge](https://img.shields.io/badge/OpenTracing-enabled-blue.svg)](http://opentracing.io)

Flaki service is a service that provides grpc and http access to [Flaki](https://github.com/cloudtrust/flaki), a unique id generator.
## Usage
### Build
Requirements:
flatbuffer


```bash
./scripts/build.sh --env <value>
```
### Launch
Launch the flaki service:
```bash
./bin/flaki_service
```
By default, when you launch the flaki service, the parameters are read from ```./conf/DEV/flaki_service.yml```.

You can load a different configuration file with:
```bash
./bin/flaki_service --config-file "path/to/file.yml".
```
### Configuration
You need to configure:
- http and grpc addresses
- the flaki ID generator
- InfluxDB
- Sentry

The flaki service will listen to the http and grpc addresses given in the configuration file.

The ID generator needs an componenet ID and a node ID. There should'nt be another instance of the generator with same componenent AND node ID. This way we can ensure the uniqueness of the generated IDs.

The metrics will be send to an Influx time series DB. The service needs the url, username, password and db name of the DB.

The errors and crashes are sent to a sentry error tracking system. We need the sentry DSN.

### grpc client


### http client

 
In influxDB, tags are indexed, so we put the correalationID as tags to speed up queries.
To query over a tag: DO NOT FORGET TO SIMPLE QUOTE the tag, otherwise it returns always empty results.
select * from "nextID-endpoint" where "correlationID" = '0';


Activation / deactivation of the modules (metrics, tracing, ....)
travis

go list ./... | grep -v /vendor/ | grep -v /client | grep -v /service/transport/flatbuffer/fb | xargs go test
TODO make travis work with multiple packages
https://github.com/mattn/goveralls/issues/20


