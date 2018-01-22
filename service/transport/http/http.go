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
)

func MakeNextIDHandler(e endpoint.Endpoint, log log.Logger, tracer opentracing.Tracer) *http_transport.Server {
	return http_transport.NewServer(e,
		decodeNextIDRequest,
		encodeNextIDResponse,
		http_transport.ServerErrorEncoder(MakeNextIDErrorHandler(log)),
		http_transport.ServerBefore(fetchCorrelationID),
		http_transport.ServerBefore(makeTracerHandler(tracer, "nextID")),
	)
}

func MakeNextValidIDHandler(e endpoint.Endpoint, log log.Logger, tracer opentracing.Tracer) *http_transport.Server {
	return http_transport.NewServer(e,
		decodeNextValidIDRequest,
		encodeNextValidIDResponse,
		http_transport.ServerErrorEncoder(MakeNextIDErrorHandler(log)),
		http_transport.ServerBefore(fetchCorrelationID),
		http_transport.ServerBefore(makeTracerHandler(tracer, "nextValidID")),
	)
}

// fetchCorrelationID read the correlation id from the http header "X-Correlation-ID".
// If the id is not zero, we put it in the context.
func fetchCorrelationID(ctx context.Context, r *http.Request) context.Context {
	var correlationID = r.Header.Get("X-Correlation-ID")
	if correlationID != "" {
		ctx = context.WithValue(ctx, "correlationID", correlationID)
	}
	return ctx
}

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
		sc.ForeachBaggageItem(func(k, v string) bool {
			fmt.Printf("key: %s, val: %s\n", k, v)
			return true
		})

		return opentracing.ContextWithSpan(r.Context(), span)
	}
}

func decodeNextIDRequest(_ context.Context, r *http.Request) (res interface{}, err error) {
	var data []byte

	data, err = ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	return fb.GetRootAsEmptyRequest(data, 0), nil
}

func encodeNextIDResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)

	var b = flatbuffers.NewBuilder(0)
	fb.NextIDReplyStart(b)
	fb.NextIDReplyAddId(b, response.(uint64))
	b.Finish(fb.NextIDReplyEnd(b))

	w.Write(b.FinishedBytes())
	return nil
}

func MakeNextIDErrorHandler(logger log.Logger) http_transport.ErrorEncoder {
	return func(ctx context.Context, err error, w http.ResponseWriter) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusInternalServerError)

		var b = flatbuffers.NewBuilder(0)
		var errStr = b.CreateString(err.Error())

		fb.NextIDReplyStart(b)
		fb.NextIDReplyAddId(b, 0)
		fb.NextIDReplyAddError(b, errStr)
		b.Finish(fb.NextValidIDReplyEnd(b))

		w.Write(b.FinishedBytes())
	}
}

func decodeNextValidIDRequest(_ context.Context, r *http.Request) (res interface{}, err error) {
	var data []byte

	data, err = ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	return fb.GetRootAsEmptyRequest(data, 0), nil
}

func encodeNextValidIDResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)

	var b = flatbuffers.NewBuilder(0)
	fb.NextValidIDReplyStart(b)
	fb.NextValidIDReplyAddId(b, response.(uint64))
	b.Finish(fb.NextValidIDReplyEnd(b))

	w.Write(b.FinishedBytes())
	return nil
}

func MakeVersion(componentName, version, environment, gitCommit string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("Component name: %s, version: %s, environment: %s, git commit: %s\n", componentName, version, environment, gitCommit)))
	}
}
