package flaki

import (
	"context"
	"io/ioutil"
	"net/http"

	"github.com/cloudtrust/flaki-service/api/fb"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	http_transport "github.com/go-kit/kit/transport/http"
	"github.com/google/flatbuffers/go"
	"github.com/pkg/errors"
)

// Flaki is the interface of the distributed unique IDs generator.
type Flaki interface {
	NextValidID(context.Context) string
}

// MakeHTTPNextIDHandler makes a HTTP handler for the NextID endpoint.
func MakeHTTPNextIDHandler(e endpoint.Endpoint, flaki Flaki, logger log.Logger) *http_transport.Server {
	return http_transport.NewServer(e,
		decodeHTTPRequest,
		encodeHTTPReply,
		http_transport.ServerErrorEncoder(httpErrorHandler),
		http_transport.ServerBefore(makeHTTPCorrelationID(flaki, logger)),
	)
}

// MakeHTTPNextValidIDHandler makes a HTTP handler for the NextValidID endpoint.
func MakeHTTPNextValidIDHandler(e endpoint.Endpoint, flaki Flaki) *http_transport.Server {
	return http_transport.NewServer(e,
		decodeHTTPRequest,
		encodeHTTPReply,
		http_transport.ServerErrorEncoder(httpErrorHandler),
		http_transport.ServerBefore(makeHTTPCorrelationID(flaki, logger)),
	)
}

// makeHTTPCorrelationID put a correlation ID in the context under the key 'correlation_id'.
// It takes the correlation ID from the http header "X-Correlation-ID", or generates a new one
// if there is no such header.
func makeHTTPCorrelationID(flaki Flaki, logger log.Logger) func(context.Context, *http.Request) context.Context {
	return func(ctx context.Context, req *http.Request) context.Context {
		// Fetch correlation header from HTTP header.
		var correlationID = req.Header.Get("X-Correlation-ID")
		if correlationID != "" {
			return context.WithValue(ctx, "correlation_id", correlationID)
		}

		// Generate correlation ID.
		correlationID = flaki.NextValidID(context.Background())

		return ctx
	}
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
