package http

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	fb "github.com/JohanDroz/flaki-service/service/transport/flatbuffer/flaki"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	http_transport "github.com/go-kit/kit/transport/http"
	"github.com/google/flatbuffers/go"
)

/*
func NewFlakiHandler(logger log.Logger, flaki flaki.Flaki) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-protobuf")
		w.Write([]byte("Test"))
	}
}*/

func MakeNextIDHandler(e endpoint.Endpoint, log log.Logger) *http_transport.Server {
	return http_transport.NewServer(e,
		decodeNextIDRequest,
		encodeNextIDResponse,
		http_transport.ServerErrorEncoder(MakeNextIDErrorHandler(log)),
	)
}

func decodeNextIDRequest(_ context.Context, r *http.Request) (res interface{}, err error) {
	var data []byte
	data, err = ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	var empty = fb.GetRootAsEmptyRequest(data, 0)

	return empty, nil
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

func MakeNextValidIDHandler(e endpoint.Endpoint, log log.Logger) *http_transport.Server {
	return http_transport.NewServer(e,
		decodeNextValidIDRequest,
		encodeNextValidIDResponse,
	)
}

func decodeNextValidIDRequest(_ context.Context, r *http.Request) (res interface{}, err error) {
	var data []byte
	data, err = ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	var empty = fb.GetRootAsEmptyRequest(data, 0)

	return empty, nil
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
