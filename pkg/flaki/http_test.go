package flaki

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/cloudtrust/flaki-service/api/fb"
	"github.com/cloudtrust/flaki-service/pkg/flaki/mock"

	"github.com/go-kit/kit/ratelimit"
	"github.com/golang/mock/gomock"
	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

func TestHTTPNextIDHandler(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockComponent = mock.NewIDGeneratorComponent(mockCtrl)

	var nextIDHandler = MakeHTTPNextIDHandler(MakeNextIDEndpoint(mockComponent))

	rand.Seed(time.Now().UnixNano())
	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var req = createFlakiRequest()
	var reply = createFlakiReply(flakiID)

	// Flatbuffer request.
	var b = flatbuffers.NewBuilder(0)
	fb.FlakiRequestStart(b)
	b.Finish(fb.FlakiRequestEnd(b))

	// HTTP request.
	var httpReq = httptest.NewRequest("POST", "http://cloudtrust.io/nextid", bytes.NewReader(b.FinishedBytes()))
	var w = httptest.NewRecorder()

	// NextID.
	mockComponent.EXPECT().NextID(context.Background(), req).Return(reply, nil).Times(1)
	nextIDHandler.ServeHTTP(w, httpReq)
	var res = w.Result()
	var body, err = ioutil.ReadAll(res.Body)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "application/octet-stream", res.Header.Get("Content-Type"))
	// Decode and check reply.
	var r = fb.GetRootAsFlakiReply(body, 0)
	assert.Equal(t, flakiID, string(r.Id()))
}

func TestHTTPNextValidIDHandler(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockComponent = mock.NewIDGeneratorComponent(mockCtrl)

	var nextValidIDHandler = MakeHTTPNextValidIDHandler(MakeNextValidIDEndpoint(mockComponent))

	rand.Seed(time.Now().UnixNano())
	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var req = createFlakiRequest()
	var reply = createFlakiReply(flakiID)

	// Flatbuffer request.
	var b = flatbuffers.NewBuilder(0)
	fb.FlakiRequestStart(b)
	b.Finish(fb.FlakiRequestEnd(b))

	// HTTP request.
	var httpReq = httptest.NewRequest("POST", "http://cloudtrust.io/nextvalidid", bytes.NewReader(b.FinishedBytes()))
	var w = httptest.NewRecorder()

	// NextValidID.
	mockComponent.EXPECT().NextValidID(context.Background(), req).Return(reply).Times(1)
	nextValidIDHandler.ServeHTTP(w, httpReq)
	var res = w.Result()
	var body, err = ioutil.ReadAll(res.Body)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "application/octet-stream", res.Header.Get("Content-Type"))
	// Decode and check reply.
	var r = fb.GetRootAsFlakiReply(body, 0)
	assert.Equal(t, flakiID, string(r.Id()))
}

func TestHTTPErrorHandler(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockComponent = mock.NewIDGeneratorComponent(mockCtrl)

	var nextIDHandler = MakeHTTPNextIDHandler(MakeNextIDEndpoint(mockComponent))

	// Flatbuffer request.
	var b = flatbuffers.NewBuilder(0)
	fb.FlakiRequestStart(b)
	b.Finish(fb.FlakiRequestEnd(b))

	// HTTP request.
	var httpReq = httptest.NewRequest("POST", "http://cloudtrust.io/nextid", bytes.NewReader(b.FinishedBytes()))
	var w = httptest.NewRecorder()
	var req = createFlakiRequest()

	// NextID.
	mockComponent.EXPECT().NextID(context.Background(), req).Return(nil, fmt.Errorf("fail")).Times(1)
	nextIDHandler.ServeHTTP(w, httpReq)
	var res = w.Result()
	var body, err = ioutil.ReadAll(res.Body)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	assert.Equal(t, "application/octet-stream", res.Header.Get("Content-Type"))
	assert.Equal(t, "fail", string(body))
}

func TestFetchHTTPCorrelationID(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockComponent = mock.NewIDGeneratorComponent(mockCtrl)

	var nextIDHandler = MakeHTTPNextIDHandler(MakeNextIDEndpoint(mockComponent))

	rand.Seed(time.Now().UnixNano())
	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", corrID)
	var req = createFlakiRequest()
	var reply = createFlakiReply(flakiID)

	// Flatbuffer request.
	var b = flatbuffers.NewBuilder(0)
	fb.FlakiRequestStart(b)
	b.Finish(fb.FlakiRequestEnd(b))

	// HTTP request.
	var httpReq = httptest.NewRequest("POST", "http://cloudtrust.io/nextid", bytes.NewReader(b.FinishedBytes()))
	httpReq.Header.Add("X-Correlation-ID", corrID)
	var w = httptest.NewRecorder()

	mockComponent.EXPECT().NextID(ctx, req).Return(reply, nil).Times(1)
	nextIDHandler.ServeHTTP(w, httpReq)
}

func TestTooManyRequests(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockComponent = mock.NewIDGeneratorComponent(mockCtrl)

	var rateLimit = 1

	var e = MakeNextIDEndpoint(mockComponent)
	e = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), rateLimit))(e)
	var h = MakeHTTPNextIDHandler(e)

	// Flatbuffer request.
	var b = flatbuffers.NewBuilder(0)
	fb.FlakiRequestStart(b)
	b.Finish(fb.FlakiRequestEnd(b))

	rand.Seed(time.Now().UnixNano())
	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var reply = createFlakiReply(flakiID)
	var req = createFlakiRequest()

	mockComponent.EXPECT().NextID(context.Background(), req).Return(reply, nil).Times(rateLimit)

	// Make too many requests, to trigger the rate limitation.
	var w *httptest.ResponseRecorder
	for i := 0; i < rateLimit+1; i++ {
		w = httptest.NewRecorder()
		var httpReq = httptest.NewRequest("POST", "http://cloudtrust.io/nextid", bytes.NewReader(b.FinishedBytes()))
		h.ServeHTTP(w, httpReq)
	}

	// Check the error returned by the rate limiter. The package ratelimit return the error
	// ErrLimited = errors.New("rate limit exceeded") when the rate is limited. In our http
	// package, we return a 429 status code when such an error arises.
	var resp = w.Result()
	var _, err = ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
}
