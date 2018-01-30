package http

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/cloudtrust/flaki-service/service/transport/flatbuffer/fb"
	"github.com/go-kit/kit/endpoint"
	http_transport "github.com/go-kit/kit/transport/http"
	"github.com/google/flatbuffers/go"
	opentracing "github.com/opentracing/opentracing-go"
)

// MakeNextIDHandler makes a HTTP handler for the NextID endpoint.
func MakeNextIDHandler(e endpoint.Endpoint, tracer opentracing.Tracer) *http_transport.Server {
	return http_transport.NewServer(e,
		decodeFlakiRequest,
		encodeFlakiReply,
		http_transport.ServerErrorEncoder(flakiErrorHandler),
		http_transport.ServerBefore(fetchCorrelationID),
	)
}

// MakeNextValidIDHandler makes a HTTP handler for the NextValidID endpoint.
func MakeNextValidIDHandler(e endpoint.Endpoint, tracer opentracing.Tracer) *http_transport.Server {
	return http_transport.NewServer(e,
		decodeFlakiRequest,
		encodeFlakiReply,
		http_transport.ServerErrorEncoder(flakiErrorHandler),
		http_transport.ServerBefore(fetchCorrelationID),
	)
}

type info struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Env     string `json:"environment"`
	Commit  string `json:"commit"`
}

// MakeVersion makes a HTTP handler that returns information about the version of the service.
func MakeVersion(componentName, version, environment, gitCommit string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var infos = info{
			Name:    componentName,
			Version: version,
			Env:     environment,
			Commit:  gitCommit,
		}

		var j, err = json.Marshal(infos)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write(j)
		}
	}
}

// fetchCorrelationID reads the correlation id from the http header "X-Correlation-ID".
// If the id is not zero, we put it in the context.
func fetchCorrelationID(ctx context.Context, r *http.Request) context.Context {
	var correlationID = r.Header.Get("X-Correlation-ID")
	if correlationID != "" {
		ctx = context.WithValue(ctx, "correlation_id", correlationID)
	}
	return ctx
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
