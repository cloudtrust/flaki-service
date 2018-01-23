package http

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	fb "github.com/cloudtrust/flaki-service/service/transport/flatbuffer/flaki"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	http_transport "github.com/go-kit/kit/transport/http"
	"github.com/google/flatbuffers/go"
	opentracing "github.com/opentracing/opentracing-go"
	opentracing_tag "github.com/opentracing/opentracing-go/ext"
)

// MakeNextIDHandler makes a HTTP handler for the NextID endpoint.
func MakeNextIDHandler(e endpoint.Endpoint, log log.Logger, tracer opentracing.Tracer) *http_transport.Server {
	return http_transport.NewServer(e,
		decodeFlakiRequest,
		encodeFlakiReply,
		http_transport.ServerErrorEncoder(flakiErrorHandler),
		http_transport.ServerBefore(fetchCorrelationID),
		http_transport.ServerBefore(makeTracerHandler(tracer, "nextID")),
	)
}

// MakeNextValidIDHandler makes a HTTP handler for the NextValidID endpoint.
func MakeNextValidIDHandler(e endpoint.Endpoint, log log.Logger, tracer opentracing.Tracer) *http_transport.Server {
	return http_transport.NewServer(e,
		decodeFlakiRequest,
		encodeFlakiReply,
		http_transport.ServerErrorEncoder(flakiErrorHandler),
		http_transport.ServerBefore(fetchCorrelationID),
		http_transport.ServerBefore(makeTracerHandler(tracer, "nextValidID")),
	)
}

// MakeVersion makes a HTTP handler that returns information about the version of the service.
func MakeVersion(componentName, version, environment, gitCommit string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("Component name: %s, version: %s, environment: %s, git commit: %s\n", componentName, version, environment, gitCommit)))
	}
}

// fetchCorrelationID reads the correlation id from the http header "X-Correlation-ID".
// If the id is not zero, we put it in the context.
func fetchCorrelationID(ctx context.Context, r *http.Request) context.Context {
	var correlationID = r.Header.Get("X-Correlation-ID")
	if correlationID != "" {
		ctx = context.WithValue(ctx, "correlation-id", correlationID)
	}
	return ctx
}

// makeTracerHandler try to extract an existing span from the HTTP headers. It it exists, we
// continue the span, if not we create a new one.
func makeTracerHandler(tracer opentracing.Tracer, operationName string) http_transport.RequestFunc {
	return func(ctx context.Context, r *http.Request) context.Context {
		var sc, err = tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))

		var span opentracing.Span
		if err != nil {
			span = tracer.StartSpan(operationName)
		} else {
			span = tracer.StartSpan(operationName, opentracing.ChildOf(sc))
		}
		defer span.Finish()

		// Set tags.
		opentracing_tag.Component.Set(span, "flaki-service")
		opentracing_tag.HTTPMethod.Set(span, operationName)
		var newctx = opentracing.ContextWithSpan(ctx, span)
		return newctx
	}
}

// decodeFlakiRequest decodes the flatbuffer flaki request.
func decodeFlakiRequest(_ context.Context, r *http.Request) (res interface{}, err error) {
	var data []byte

	data, err = ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	return fb.GetRootAsEmptyRequest(data, 0), nil
}

// encodeFlakiReply encodes the flatbuffer flaki reply.
func encodeFlakiReply(_ context.Context, w http.ResponseWriter, res interface{}) error {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)

	var b = flatbuffers.NewBuilder(0)
	var id = b.CreateString(res.(string))

	fb.FlakiReplyStart(b)
	fb.FlakiReplyAddId(b, id)
	b.Finish(fb.FlakiReplyEnd(b))

	w.Write(b.FinishedBytes())
	return nil
}

// flakiErrorHandler encodes the flatbuffer flaki reply when there is an error.
func flakiErrorHandler(ctx context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusInternalServerError)

	var b = flatbuffers.NewBuilder(0)
	var errStr = b.CreateString(err.Error())

	fb.FlakiReplyStart(b)
	fb.FlakiReplyAddId(b, 0)
	fb.FlakiReplyAddError(b, errStr)
	b.Finish(fb.FlakiReplyEnd(b))

	w.Write(b.FinishedBytes())
}
