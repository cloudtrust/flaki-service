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

	rand.Seed(time.Now().UnixNano())
	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var nextIDHandler = MakeHTTPNextIDHandler(MakeNextIDEndpoint(mockComponent))

	// Flatbuffer request.
	var b = flatbuffers.NewBuilder(0)
	fb.EmptyRequestStart(b)
	b.Finish(fb.EmptyRequestEnd(b))

	// HTTP request.
	var req = httptest.NewRequest("POST", "http://cloudtrust.io/nextid", bytes.NewReader(b.FinishedBytes()))
	var w = httptest.NewRecorder()

	// NextID.
	mockComponent.EXPECT().NextID(context.Background()).Return(flakiID, nil).Times(1)
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
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockComponent = mock.NewComponent(mockCtrl)

	rand.Seed(time.Now().UnixNano())
	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var nextValidIDHandler = MakeHTTPNextValidIDHandler(MakeNextValidIDEndpoint(mockComponent))

	// Flatbuffer request.
	var b = flatbuffers.NewBuilder(0)
	fb.EmptyRequestStart(b)
	b.Finish(fb.EmptyRequestEnd(b))

	// HTTP request.
	var req = httptest.NewRequest("POST", "http://cloudtrust.io/nextvalidid", bytes.NewReader(b.FinishedBytes()))
	var w = httptest.NewRecorder()

	// NextValidID.
	mockComponent.EXPECT().NextValidID(context.Background()).Return(flakiID).Times(1)
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
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockComponent = mock.NewComponent(mockCtrl)

	var nextIDHandler = MakeHTTPNextIDHandler(MakeNextIDEndpoint(mockComponent))

	// Flatbuffer request.
	var b = flatbuffers.NewBuilder(0)
	fb.EmptyRequestStart(b)
	b.Finish(fb.EmptyRequestEnd(b))

	// HTTP request.
	var req = httptest.NewRequest("POST", "http://cloudtrust.io/nextid", bytes.NewReader(b.FinishedBytes()))
	var w = httptest.NewRecorder()

	// NextID.
	mockComponent.EXPECT().NextID(context.Background()).Return("", fmt.Errorf("fail")).Times(1)
	nextIDHandler.ServeHTTP(w, req)
	var resp = w.Result()
	var body, err = ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.Equal(t, "application/octet-stream", resp.Header.Get("Content-Type"))
	// Decode and check reply.
	var r = fb.GetRootAsFlakiReply(body, 0)
	assert.Zero(t, string(r.Id()))
	assert.NotZero(t, string(r.Error()))
}

func TestFetchHTTPCorrelationID(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockComponent = mock.NewComponent(mockCtrl)

	// Context with correlation ID.
	rand.Seed(time.Now().UnixNano())
	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", corrID)

	var nextIDHandler = MakeHTTPNextIDHandler(MakeNextIDEndpoint(mockComponent))

	// Flatbuffer request.
	var b = flatbuffers.NewBuilder(0)
	fb.EmptyRequestStart(b)
	b.Finish(fb.EmptyRequestEnd(b))

	// HTTP request.
	var req = httptest.NewRequest("POST", "http://cloudtrust.io/nextid", bytes.NewReader(b.FinishedBytes()))
	req.Header.Add("X-Correlation-ID", corrID)
	var w = httptest.NewRecorder()

	mockComponent.EXPECT().NextID(ctx).Return(flakiID, nil).Times(1)
	nextIDHandler.ServeHTTP(w, req)
}
