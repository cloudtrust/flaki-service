package flaki

import (
	"bytes"
	"context"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/cloudtrust/flaki-service/pkg/flaki/flatbuffer/fb"
	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/stretchr/testify/assert"
)

func TestHTTPNextIDHandler(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var nextIDHandler = MakeHTTPNextIDHandler(MakeMockEndpoint(flakiID, false))

	// Flatbuffer request.
	var b = flatbuffers.NewBuilder(0)
	fb.EmptyRequestStart(b)
	b.Finish(fb.EmptyRequestEnd(b))

	// HTTP request.
	var req = httptest.NewRequest("POST", "http://cloudtrust.io/nextid", bytes.NewReader(b.FinishedBytes()))
	var w = httptest.NewRecorder()

	// NextID.
	nextIDHandler.ServeHTTP(w, req)
	var resp = w.Result()
	var body, err = ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/octet-stream", resp.Header.Get("Content-Type"))
	// Decode and check reply.
	var r = fb.GetRootAsFlakiReply(body, 0)
	assert.Equal(t, flakiID, string(r.Id()))
	assert.Zero(t, string(r.Error()))
}
func TestHTTPNextValidIDHandler(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var nextValidIDHandler = MakeHTTPNextValidIDHandler(MakeMockEndpoint(flakiID, false))

	// Flatbuffer request.
	var b = flatbuffers.NewBuilder(0)
	fb.EmptyRequestStart(b)
	b.Finish(fb.EmptyRequestEnd(b))

	// HTTP request.
	var req = httptest.NewRequest("POST", "http://cloudtrust.io/nextvalidid", bytes.NewReader(b.FinishedBytes()))
	var w = httptest.NewRecorder()

	// NextValidID.
	nextValidIDHandler.ServeHTTP(w, req)
	var resp = w.Result()
	var body, err = ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/octet-stream", resp.Header.Get("Content-Type"))
	// Decode and check reply.
	var r = fb.GetRootAsFlakiReply(body, 0)
	assert.Equal(t, flakiID, string(r.Id()))
	assert.Zero(t, string(r.Error()))
}

func TestHTTPErrorHandler(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var nextIDHandler = MakeHTTPNextIDHandler(MakeMockEndpoint(flakiID, true))

	// Flatbuffer request.
	var b = flatbuffers.NewBuilder(0)
	fb.EmptyRequestStart(b)
	b.Finish(fb.EmptyRequestEnd(b))

	// HTTP request.
	var req = httptest.NewRequest("POST", "http://cloudtrust.io/nextid", bytes.NewReader(b.FinishedBytes()))
	var w = httptest.NewRecorder()

	// NextID.
	nextIDHandler.ServeHTTP(w, req)
	var resp = w.Result()
	var body, err = ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.Equal(t, "application/octet-stream", resp.Header.Get("Content-Type"))
	// Decode and check reply.
	var r = fb.GetRootAsFlakiReply(body, 0)
	assert.Zero(t, string(r.Id()))
	assert.Equal(t, "fail", string(r.Error()))
}

func TestFetchHTTPCorrelationID(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var endpoint = func(ctx context.Context, request interface{}) (response interface{}, err error) {
		var id = ctx.Value("correlation_id")
		assert.NotNil(t, id)
		assert.Equal(t, corrID, id.(string))

		return "", nil
	}
	var nextIDHandler = MakeHTTPNextIDHandler(endpoint)

	// Flatbuffer request.
	var b = flatbuffers.NewBuilder(0)
	fb.EmptyRequestStart(b)
	b.Finish(fb.EmptyRequestEnd(b))

	// HTTP request.
	var req = httptest.NewRequest("POST", "http://cloudtrust.io/nextid", bytes.NewReader(b.FinishedBytes()))
	req.Header.Add("X-Correlation-ID", corrID)
	var w = httptest.NewRecorder()

	nextIDHandler.ServeHTTP(w, req)
}
