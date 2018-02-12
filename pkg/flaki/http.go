package flaki

import (
	"context"
	"io/ioutil"
	"net/http"

	"github.com/cloudtrust/flaki-service/pkg/flaki/flatbuffer/fb"
	"github.com/go-kit/kit/endpoint"
	http_transport "github.com/go-kit/kit/transport/http"
	"github.com/google/flatbuffers/go"
)

// MakeHTTPNextIDHandler makes a HTTP handler for the NextID endpoint.
func MakeHTTPNextIDHandler(e endpoint.Endpoint) *http_transport.Server {
	return http_transport.NewServer(e,
		decodeHTTPRequest,
		encodeHTTPReply,
		http_transport.ServerErrorEncoder(httpErrorHandler),
		http_transport.ServerBefore(fetchHTTPCorrelationID),
	)
}

// MakeHTTPNextValidIDHandler makes a HTTP handler for the NextValidID endpoint.
func MakeHTTPNextValidIDHandler(e endpoint.Endpoint) *http_transport.Server {
	return http_transport.NewServer(e,
		decodeHTTPRequest,
		encodeHTTPReply,
		http_transport.ServerErrorEncoder(httpErrorHandler),
		http_transport.ServerBefore(fetchHTTPCorrelationID),
	)
}

// fetchHTTPCorrelationID reads the correlation id from the http header "X-Correlation-ID".
// If the id is not zero, we put it in the context.
func fetchHTTPCorrelationID(ctx context.Context, r *http.Request) context.Context {
	var correlationID = r.Header.Get("X-Correlation-ID")
	if correlationID != "" {
		ctx = context.WithValue(ctx, "correlation_id", correlationID)
	}
	return ctx
}

// decodeHTTPRequest decodes the flatbuffer flaki request.
func decodeHTTPRequest(_ context.Context, req *http.Request) (interface{}, error) {
	var data, err = ioutil.ReadAll(req.Body)

	if err != nil {
		return nil, err
	}

	return fb.GetRootAsEmptyRequest(data, 0), nil
}

// encodeHTTPReply encodes the flatbuffer flaki reply.
func encodeHTTPReply(_ context.Context, w http.ResponseWriter, res interface{}) error {
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

// httpErrorHandler encodes the flatbuffer flaki reply when there is an error.
func httpErrorHandler(ctx context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusInternalServerError)

	var b = flatbuffers.NewBuilder(0)
	var errStr = b.CreateString(err.Error())

	fb.FlakiReplyStart(b)
	fb.FlakiReplyAddError(b, errStr)
	b.Finish(fb.FlakiReplyEnd(b))

	w.Write(b.FinishedBytes())
}
