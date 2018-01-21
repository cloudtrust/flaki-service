package main

import (
	"bytes"
	fb "github.com/JohanDroz/flaki-service/service/transport/flatbuffer/flaki"
	"github.com/go-kit/kit/log"
	"github.com/google/flatbuffers/go"
	"io/ioutil"
	"net/http"
	"os"
)

const (
	address = "http://localhost:8888"
)

func main() {

	// Logger.
	var logger = log.NewLogfmtLogger(os.Stdout)
	{
		logger = log.With(logger, "time", log.DefaultTimestampUTC, "caller", log.DefaultCaller)
		defer logger.Log("msg", "Goodbye")
	}
	logger = log.With(logger, "transport", "http")

	// Empty request
	var b = flatbuffers.NewBuilder(0)
	fb.EmptyRequestStart(b)
	b.Finish(fb.EmptyRequestEnd(b))

	// http NextID
	var httpNextIDResp *http.Response
	{
		var err error
		httpNextIDResp, err = http.Post(address+"/nextid", "application/octet-stream", bytes.NewReader(b.FinishedBytes()))
		if err != nil {
			logger.Log("error", err)
		}
		defer httpNextIDResp.Body.Close()

		// Read flatbuffer reply.
		var data []byte
		data, err = ioutil.ReadAll(httpNextIDResp.Body)
		if err != nil {
			logger.Log("error", err)
		}

		var nextIDReply = fb.GetRootAsNextIDReply(data, 0)
		logger.Log("endpoint", "nextID", "id", nextIDReply.Id(), "error", nextIDReply.Error())
	}

	// http NextValidID
	var httpNextValidIDResp *http.Response
	{
		var err error
		httpNextValidIDResp, err = http.Post(address+"/nextvalidid", "application/octet-stream", bytes.NewReader(b.FinishedBytes()))
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
	}
}
