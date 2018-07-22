package flaki

import (
	"context"
	"io/ioutil"
	"net/http"

	"github.com/cloudtrust/flaki-service/api/fb"
	"github.com/go-kit/kit/endpoint"
	http_transport "github.com/go-kit/kit/transport/http"
	"github.com/google/flatbuffers/go"
	"github.com/pkg/errors"
)

// MakeHTTPNextIDHandler makes a HTTP handler for the NextID endpoint.
func MakeHTTPNextIDHandler(e endpoint.Endpoint) *http_transport.Server {
	return http_transport.NewServer(e,
		decodeHTTPRequest,
		encodeHTTPReply,
		http_transport.ServerErrorEncoder(httpErrorHandler),
	)
}

// MakeHTTPNextValidIDHandler makes a HTTP handler for the NextValidID endpoint.
func MakeHTTPNextValidIDHandler(e endpoint.Endpoint) *http_transport.Server {
	return http_transport.NewServer(e,
		decodeHTTPRequest,
		encodeHTTPReply,
		http_transport.ServerErrorEncoder(httpErrorHandler),
	)
}

// decodeHTTPRequest decodes the flatbuffer flaki request.
func decodeHTTPRequest(_ context.Context, req *http.Request) (interface{}, error) {
	var data, err = ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, errors.Wrap(err, "could not decode HTTP request")
	}

	return fb.GetRootAsFlakiRequest(data, 0), nil
}

// encodeHTTPReply encodes the flatbuffer flaki reply.
func encodeHTTPReply(_ context.Context, w http.ResponseWriter, rep interface{}) error {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)

	var reply = rep.(*fb.FlakiReply)

	var b = flatbuffers.NewBuilder(0)
	var str = b.CreateString(string(reply.Id()))

	fb.FlakiReplyStart(b)
	fb.FlakiReplyAddId(b, str)
	b.Finish(fb.FlakiReplyEnd(b))

	w.Write(b.FinishedBytes())
	return nil
}

// httpErrorHandler encodes the flatbuffer flaki reply when there is an error.
func httpErrorHandler(ctx context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/octet-stream")

	switch err.Error() {
	case "rate limit exceeded":
		w.WriteHeader(http.StatusTooManyRequests)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.Write([]byte(err.Error()))
}
