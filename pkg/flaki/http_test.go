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

	"github.com/cloudtrust/flaki-service/pkg/flaki/flatbuffer/fb"
	"github.com/cloudtrust/flaki-service/pkg/flaki/mock"
	"github.com/golang/mock/gomock"
	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/stretchr/testify/assert"
)

func TestHTTPNextIDHandler(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockComponent = mock.NewComponent(mockCtrl)

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
	var mockComponent = mock.NewComponent(mockCtrl)

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
	var mockComponent = mock.NewComponent(mockCtrl)

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
	var mockComponent = mock.NewComponent(mockCtrl)

	var nextIDHandler = MakeHTTPNextIDHandler(MakeNextIDEndpoint(mockComponent))

	rand.Seed(time.Now().UnixNano())
	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), CorrelationIDKey, corrID)
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
