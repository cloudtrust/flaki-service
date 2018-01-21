package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"strconv"

	fb "github.com/JohanDroz/flaki-service/service/transport/flatbuffer/flaki"
	"github.com/go-kit/kit/log"
	"github.com/google/flatbuffers/go"
	opentracing "github.com/opentracing/opentracing-go"
	opentracing_log "github.com/opentracing/opentracing-go/log"
	jaeger_client "github.com/uber/jaeger-client-go/config"
)

const (
	address = "127.0.0.1:8888"
)

func main() {

	// Logger.
	var logger = log.NewLogfmtLogger(os.Stdout)
	{
		logger = log.With(logger, "time", log.DefaultTimestampUTC, "caller", log.DefaultCaller)
		defer logger.Log("msg", "Goodbye")
	}
	logger = log.With(logger, "transport", "http")

	var jaegerConfig = jaeger_client.Configuration{
		Sampler: &jaeger_client.SamplerConfig{
			Type:              "const",
			Param:             1,
			SamplingServerURL: "http://127.0.0.1:5775/",
		},
		Reporter: &jaeger_client.ReporterConfig{
			LogSpans:            false,
			BufferFlushInterval: 1000 * time.Millisecond,
		},
	}

	// Jaeger client
	var tracer opentracing.Tracer
	{
		var logger = log.With(logger, "component", "jaeger")
		var closer io.Closer
		var err error

		tracer, closer, err = jaegerConfig.New("flaki_http")
		if err != nil {
			logger.Log("error", err)
			return
		}
		defer closer.Close()
	}

	var id, err = nextID(logger, tracer)
	fmt.Printf("ID: %d, err: %v", id, err)
}

func nextID(logger log.Logger, tracer opentracing.Tracer) (uint64, error) {
	// Empty request
	var b = flatbuffers.NewBuilder(0)
	fb.EmptyRequestStart(b)
	b.Finish(fb.EmptyRequestEnd(b))

	// http NextID
	var httpNextIDResp *http.Response
	{
		var err error
		var req *http.Request
		var url = fmt.Sprintf("http://%s/nextid", address)

		req, err = http.NewRequest("POST", url, bytes.NewReader(b.FinishedBytes()))
		if err != nil {
			logger.Log("error", err)
		}

		var span = tracer.StartSpan("nexitJDR")
		defer span.Finish()
		span.LogFields(
			opentracing_log.String("operation", "nexid"),
			opentracing_log.String("microservice_level", "transport"),
		)
		span = span.SetBaggageItem("opentracing-baguage", "my_baguage")
		fmt.Printf("Span context: %s\n", span.Context())

		var carrier = opentracing.HTTPHeadersCarrier(req.Header)
		fmt.Printf("%s", carrier)
		tracer.Inject(span.Context(), opentracing.HTTPHeaders, carrier)

		req.Header.Set("Content-Type", "application/octet-stream")
		//req.Header.Set("jaeger-baggage", "11")

		var corrID uint64 = 10 //rand.Uint64()

		req.Header.Set("X-Correlation-ID", strconv.FormatUint(corrID, 10))
		httpNextIDResp, err = http.DefaultClient.Do(req)

		if err != nil {
			logger.Log("error", err)
		}
		defer httpNextIDResp.Body.Close()

		// Read correlation id from reply.
		/*var corrIDResp uint64
		{
			corrIDResp, err = strconv.ParseUint(httpNextIDResp.Header.Get("X-Correlation-ID"), 10, 64)
			if err != nil {
				return 0, err
			}
			if corrID != corrIDResp {
				var err = fmt.Errorf("Wrong correlation id from response")
				logger.Log("error", err)
				return 0, err
			}
		}*/

		// Read flatbuffer reply.
		var data []byte
		data, err = ioutil.ReadAll(httpNextIDResp.Body)
		if err != nil {
			logger.Log("error", err)
		}

		var nextIDReply = fb.GetRootAsNextIDReply(data, 0)
		logger.Log("endpoint", "nextID", "id", nextIDReply.Id(), "error", nextIDReply.Error())

		return nextIDReply.Id(), nil
	}
}

func nextValidID(logger log.Logger) uint64 {
	// Empty request
	var b = flatbuffers.NewBuilder(0)
	fb.EmptyRequestStart(b)
	b.Finish(fb.EmptyRequestEnd(b))

	// http NextValidID
	var httpNextValidIDResp *http.Response
	{
		var err error
		httpNextValidIDResp, err = http.Post(fmt.Sprintf("http://%s/nextvalidid", address), "application/octet-stream", bytes.NewReader(b.FinishedBytes()))
		if err != nil {
			logger.Log("error", err)
		}
		defer httpNextValidIDResp.Body.Close()

		// Read flatbuffer reply.
		var data []byte
		data, err = ioutil.ReadAll(httpNextValidIDResp.Body)
		if err != nil {
			logger.Log("error", err)
		}

		var nextValidIDReply = fb.GetRootAsNextValidIDReply(data, 0)
		logger.Log("endpoint", "nextValidID", "id", nextValidIDReply.Id())
		return nextValidIDReply.Id()
	}
}
