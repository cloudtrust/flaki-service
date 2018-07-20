package health

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	http_transport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

// MakeHealthCheckHandler make an HTTP handler for an HealthCheck endpoint.
func MakeHealthCheckHandler(e endpoint.Endpoint) *http_transport.Server {
	return http_transport.NewServer(e,
		decodeHealthCheckRequest,
		encodeHealthCheckReply,
		http_transport.ServerErrorEncoder(healthCheckErrorHandler),
		http_transport.ServerBefore(fetchParameters),
	)
}

// fetchHTTPParameters gets the HTTP parameters 'module', 'healthcheck', and 'nocache'.
// They define which health check to execute, e.g. to ping influx, 'module' = influx and
// 'healthcheck' = ping.
func fetchParameters(ctx context.Context, req *http.Request) context.Context {
	// Fetch module and healthcheck name from URL path
	var m = mux.Vars(req)
	for _, key := range []string{"module", "healthcheck"} {
		ctx = context.WithValue(ctx, key, m[key])
	}

	// Fetch nocache URL param
	ctx = context.WithValue(ctx, "nocache", req.URL.Query().Get("nocache"))

	return ctx
}

// decodeHealthCheckRequest decodes the health check request.
func decodeHealthCheckRequest(_ context.Context, r *http.Request) (res interface{}, err error) {
	return nil, nil
}

// encodeHealthCheckReply encodes the health check reply.
func encodeHealthCheckReply(_ context.Context, w http.ResponseWriter, rep interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	var data, ok = rep.(json.RawMessage)

	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}

	return nil
}

// healthCheckErrorHandler encodes the health check reply when there is an error.
func healthCheckErrorHandler(ctx context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	switch err.Error() {
	case "rate limit exceeded":
		w.WriteHeader(http.StatusTooManyRequests)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}

	// Write error.
	var reply, _ = json.MarshalIndent(map[string]string{"error": err.Error()}, "", "  ")
	w.Write(reply)
}
