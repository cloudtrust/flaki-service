package flaki

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/cloudtrust/flaki-service/pkg/flaki/flatbuffer/fb"
	"github.com/cloudtrust/flaki-service/pkg/flaki/mock"
	"github.com/golang/mock/gomock"
	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/stretchr/testify/assert"
)

func TestNewComponent(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockModule = mock.NewModule(mockCtrl)

	var c = NewComponent(mockModule)

	rand.Seed(time.Now().UnixNano())
	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var req = createFlakiRequest()

	// NextID.
	mockModule.EXPECT().NextID(context.Background()).Return(flakiID, nil).Times(1)
	var reply, err = c.NextID(context.Background(), req)
	assert.Nil(t, err)
	assert.Equal(t, flakiID, string(reply.Id()))

	// NextID error.
	mockModule.EXPECT().NextID(context.Background()).Return("", fmt.Errorf("fail")).Times(1)
	reply, err = c.NextID(context.Background(), req)
	assert.NotNil(t, err)
	assert.Nil(t, reply)

	// NextValidID.
	mockModule.EXPECT().NextValidID(context.Background()).Return(flakiID).Times(1)
	reply = c.NextValidID(context.Background(), req)
	assert.Equal(t, flakiID, string(reply.Id()))
}

func createFlakiRequest() *fb.FlakiRequest {
	var b = flatbuffers.NewBuilder(0)

	fb.FlakiRequestStart(b)
	b.Finish(fb.FlakiRequestEnd(b))

	return fb.GetRootAsFlakiRequest(b.FinishedBytes(), 0)
}

func createFlakiReply(id string) *fb.FlakiReply {
	var b = flatbuffers.NewBuilder(0)
	var str = b.CreateString(id)

	fb.FlakiReplyStart(b)
	fb.FlakiReplyAddId(b, str)
	b.Finish(fb.FlakiReplyEnd(b))

	return fb.GetRootAsFlakiReply(b.FinishedBytes(), 0)
}
