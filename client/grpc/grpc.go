package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	fb "github.com/JohanDroz/flaki-service/service/transport/flatbuffer/flaki"
	"github.com/go-kit/kit/log"
	"github.com/google/flatbuffers/go"
	opentracing "github.com/opentracing/opentracing-go"
	jaeger_client "github.com/uber/jaeger-client-go/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	address = "127.0.0.1:5555"
)

func main() {

	// Logger.
	var logger = log.NewLogfmtLogger(os.Stdout)
	{
		logger = log.With(logger, "time", log.DefaultTimestampUTC, "caller", log.DefaultCaller)
		defer logger.Log("msg", "Goodbye")
	}
	logger = log.With(logger, "transport", "grpc")
	var jaegerConfig = jaeger_client.Configuration{
		Sampler: &jaeger_client.SamplerConfig{
			Type:              "const",
			Param:             1,
			SamplingServerURL: "http://127.0.0.1:5775/",
		},
		Reporter: &jaeger_client.ReporterConfig{
			LogSpans:            false,
			BufferFlushInterval: 1000 * time.Millisecond,
		},
	}

	// Jaeger client
	var tracer opentracing.Tracer
	{
		var logger = log.With(logger, "component", "jaeger")
		var closer io.Closer
		var err error

		tracer, closer, err = jaegerConfig.New("flaki_http")
		if err != nil {
			logger.Log("error", err)
			return
		}
		defer closer.Close()
	}

	// Set up a connection to the server.
	var clienConn *grpc.ClientConn
	{
		var err error
		clienConn, err = grpc.Dial(address, grpc.WithInsecure(), grpc.WithCodec(flatbuffers.FlatbuffersCodec{}))
		if err != nil {
			logger.Log("error", err)
		}
		defer clienConn.Close()
	}

	var flakiClient = fb.NewFlakiClient(clienConn)

	// Empty request
	var b = flatbuffers.NewBuilder(0)
	fb.EmptyRequestStart(b)
	b.Finish(fb.EmptyRequestEnd(b))

	var span = tracer.StartSpan("fuck")
	defer span.Finish()
	span = span.SetBaggageItem("myBaggage", "myBaggageValue")

	var m = make(opentracing.TextMapCarrier)

	var err = tracer.Inject(span.Context(), opentracing.TextMap, m)
	if err != nil {
		logger.Log("error", err)
		return
	}
	fmt.Printf("map: %s\n", m)
	fmt.Printf("span: %s\n", span)

	m.Set("id", "0")
	var md = metadata.New()

	// gRPC NextID
	var nextIDReply *fb.NextIDReply
	{
		var err error
		var ctx = metadata.NewOutgoingContext(opentracing.ContextWithSpan(context.Background(), span), md)
		nextIDReply, err = flakiClient.NextID(ctx, b)
		if err != nil {
			logger.Log("error", err)
			return
		}
		logger.Log("endpoint", "nextID", "id", nextIDReply.Id(), "error", nextIDReply.Error())
	}

	// gRPC NextValidID
	var nextValidIDReply *fb.NextValidIDReply
	{
		var err error
		var ctx = metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{"id": strconv.FormatUint(1, 10)}))
		nextValidIDReply, err = flakiClient.NextValidID(ctx, b)
		if err != nil {
			logger.Log("error", err)
			return
		}
		logger.Log("endpoint", "nextValidID", "id", nextValidIDReply.Id())
	}
}

/*
func NewGRPCClient(conn *grpc.ClientConn) flaki_component.Service {

	var nextIDEndpoint endpoint.Endpoint
	{
		nextIDEndpoint = grpc_transport.NewClient(
			conn,
			"flaki.Flaki",
			"NextID",
			encodeNextIDRequest,
			decodeNextIDResponse,
			fb.NextIDReply{},
		).Endpoint()
	}

	var nextValidIDEndpoint endpoint.Endpoint
	{
		nextValidIDEndpoint = grpc_transport.NewClient(
			conn,
			"flaki.Flaki",
			"NextValidID",
			encodeNextValidIDRequest,
			decodeNextValidIDResponse,
			fb.NextValidIDReply{},
		).Endpoint()
	}

	return &flaki_endpoint.Endpoints{
		NextIDEndpoint:      nextIDEndpoint,
		NextValidIDEndpoint: nextValidIDEndpoint,
	}
}

func encodeNextIDRequest(_ context.Context, req interface{}) (interface{}, error) {
	return req, nil
}

func encodeNextValidIDRequest(_ context.Context, req interface{}) (interface{}, error) {
	return req, nil
}
func decodeNextIDResponse(_ context.Context, req interface{}) (interface{}, error) {
	panic("decodeNextIDResponse")

	return req, nil
}

func decodeNextValidIDResponse(_ context.Context, req interface{}) (interface{}, error) {
	panic("decodeNextValidIDResponse")

	return req, nil
}*/
