package main

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/cloudtrust/flaki-service/service/transport/flatbuffer/fb"
	"github.com/go-kit/kit/log"
	"github.com/google/flatbuffers/go"
	opentracing "github.com/opentracing/opentracing-go"
	opentracing_tag "github.com/opentracing/opentracing-go/ext"
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

	// Jaeger tracer config.
	var jaegerConfig = jaeger_client.Configuration{
		Sampler: &jaeger_client.SamplerConfig{
			Type:              "const",
			Param:             1,
			SamplingServerURL: "http://127.0.0.1:5775",
		},
		Reporter: &jaeger_client.ReporterConfig{
			LogSpans:            false,
			BufferFlushInterval: 1000 * time.Millisecond,
		},
	}

	// Jaeger client.
	var tracer opentracing.Tracer
	{
		var logger = log.With(logger, "component", "jaeger")
		var closer io.Closer
		var err error

		tracer, closer, err = jaegerConfig.New("flaki-service")
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

	// Client.
	var flakiClient = fb.NewFlakiClient(clienConn)

	var span = tracer.StartSpan("grpc_client")
	opentracing_tag.SpanKindRPCClient.Set(span)
	defer span.Finish()

	nextID(flakiClient, logger, tracer, span)
	nextValidID(flakiClient, logger, tracer, span)
}

func nextID(client fb.FlakiClient, logger log.Logger, tracer opentracing.Tracer, parentSpan opentracing.Span) {
	// NextID.
	var b = flatbuffers.NewBuilder(0)
	fb.EmptyRequestStart(b)
	b.Finish(fb.EmptyRequestEnd(b))

	var span = tracer.StartSpan("grpc_client_nextid", opentracing.ChildOf(parentSpan.Context()))
	opentracing_tag.SpanKindRPCClient.Set(span)
	defer span.Finish()

	// Propagate the opentracing span.
	var carrier = make(opentracing.TextMapCarrier)
	var err = tracer.Inject(span.Context(), opentracing.TextMap, carrier)
	if err != nil {
		logger.Log("error", err)
		return
	}

	var md = metadata.New(carrier)
	var correlationIDMD = metadata.New(map[string]string{"correlation_id": "1"})

	// grpc NextID
	var nextIDreply *fb.FlakiReply
	{
		var err error
		var ctx = metadata.NewOutgoingContext(opentracing.ContextWithSpan(context.Background(), span), metadata.Join(md, correlationIDMD))
		nextIDreply, err = client.NextID(ctx, b)
		if err != nil {
			logger.Log("error", err)
			return
		}
		logger.Log("endpoint", "nextID", "id", nextIDreply.Id(), "error", nextIDreply.Error())
	}
}

func nextValidID(client fb.FlakiClient, logger log.Logger, tracer opentracing.Tracer, parentSpan opentracing.Span) {
	// NextValidID.
	var b = flatbuffers.NewBuilder(0)
	fb.EmptyRequestStart(b)
	b.Finish(fb.EmptyRequestEnd(b))

	var span = tracer.StartSpan("grpc_client_nextvalidid", opentracing.ChildOf(parentSpan.Context()))
	opentracing_tag.SpanKindRPCClient.Set(span)
	defer span.Finish()

	// Propagate the opentracing span.
	var carrier = make(opentracing.TextMapCarrier)
	var err = tracer.Inject(span.Context(), opentracing.TextMap, carrier)
	if err != nil {
		logger.Log("error", err)
		return
	}

	var md = metadata.New(carrier)
	var correlationIDMD = metadata.New(map[string]string{"correlation_id": "2"})

	// grpc NextValidID
	var nextValidIDreply *fb.FlakiReply
	{
		var err error
		var ctx = metadata.NewOutgoingContext(context.Background(), metadata.Join(md, correlationIDMD))
		nextValidIDreply, err = client.NextValidID(ctx, b)
		if err != nil {
			logger.Log("error", err)
			return
		}
		logger.Log("endpoint", "nextValidID", "id", nextValidIDreply.Id(), "error", nextValidIDreply.Error())
	}
}
