// This is an example of a HTTP client that request IDs to the flaki-service.
// Note that for the client side tracing to work, you need to have a jaeger agent
// listening to "127.0.0.1:5775". If this is not the case, the root span will be lost,
// but all other spans will be handled correctly.
package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/cloudtrust/flaki-service/api/fb"
	"github.com/go-kit/kit/log"
	"github.com/google/flatbuffers/go"
	opentracing "github.com/opentracing/opentracing-go"
	otag "github.com/opentracing/opentracing-go/ext"
	"github.com/spf13/pflag"
	jaeger_client "github.com/uber/jaeger-client-go/config"
)

var (
	host = pflag.String("host", "127.0.0.1", "flaki service host")
	port = pflag.String("port", "8888", "flaki service port")
)

func main() {
	// Configuration flags.
	pflag.Parse()

	// Logger.
	var logger = log.NewLogfmtLogger(os.Stdout)
	{
		logger = log.With(logger, "time", log.DefaultTimestampUTC, "caller", log.DefaultCaller)
		defer logger.Log("msg", "Goodbye")
	}
	logger = log.With(logger, "transport", "http")

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

		tracer, closer, err = jaegerConfig.New("flaki-client")
		if err != nil {
			logger.Log("error", err)
			return
		}
		defer closer.Close()
	}

	nextID(logger, tracer)
	nextValidID(logger, tracer)
}

func nextID(logger log.Logger, tracer opentracing.Tracer) {
	// NextID.
	var b = flatbuffers.NewBuilder(0)
	fb.FlakiRequestStart(b)
	b.Finish(fb.FlakiRequestEnd(b))

	var span = tracer.StartSpan("http")
	otag.HTTPMethod.Set(span, "http-client")
	defer span.Finish()

	// http NextID
	var httpNextIDResp *http.Response
	{
		var err error
		var req *http.Request
		var url = fmt.Sprintf("http://%s:%s/nextid", *host, *port)

		req, err = http.NewRequest("POST", url, bytes.NewReader(b.FinishedBytes()))
		if err != nil {
			logger.Log("error", err)
			return
		}

		var carrier = opentracing.HTTPHeadersCarrier(req.Header)
		tracer.Inject(span.Context(), opentracing.HTTPHeaders, carrier)

		req.Header.Set("Content-Type", "application/octet-stream")

		req.Header.Set("X-Correlation-ID", "1")
		httpNextIDResp, err = http.DefaultClient.Do(req)

		if err != nil {
			logger.Log("error", err)
			return
		}
		defer httpNextIDResp.Body.Close()

		// Read flatbuffer reply.
		var data []byte
		data, err = ioutil.ReadAll(httpNextIDResp.Body)
		if err != nil {
			logger.Log("error", err)
			return
		}

		if httpNextIDResp.StatusCode != 200 {
			logger.Log("error", string(data))
		} else {
			var reply = fb.GetRootAsFlakiReply(data, 0)
			logger.Log("endpoint", "nextValidID", "id", reply.Id())
		}
	}
}

func nextValidID(logger log.Logger, tracer opentracing.Tracer) {
	// NextID.
	var b = flatbuffers.NewBuilder(0)
	fb.FlakiRequestStart(b)
	b.Finish(fb.FlakiRequestEnd(b))

	var span = tracer.StartSpan("http")
	otag.HTTPMethod.Set(span, "http-client")
	defer span.Finish()

	// http NextValidID
	var httpNextValidIDResp *http.Response
	{
		var err error
		var req *http.Request
		var url = fmt.Sprintf("http://%s:%s/nextvalidid", *host, *port)

		req, err = http.NewRequest("POST", url, bytes.NewReader(b.FinishedBytes()))
		if err != nil {
			logger.Log("error", err)
			return
		}

		var carrier = opentracing.HTTPHeadersCarrier(req.Header)
		tracer.Inject(span.Context(), opentracing.HTTPHeaders, carrier)

		req.Header.Set("Content-Type", "application/octet-stream")

		req.Header.Set("X-Correlation-ID", "2")
		httpNextValidIDResp, err = http.DefaultClient.Do(req)

		if err != nil {
			logger.Log("error", err)
			return
		}

		// Read flatbuffer reply.
		var data []byte
		data, err = ioutil.ReadAll(httpNextValidIDResp.Body)
		if err != nil {
			logger.Log("error", err)
			return
		}

		if httpNextValidIDResp.StatusCode != 200 {
			logger.Log("error", string(data))
		} else {
			var reply = fb.GetRootAsFlakiReply(data, 0)
			logger.Log("endpoint", "nextValidID", "id", reply.Id())
		}
	}
}
