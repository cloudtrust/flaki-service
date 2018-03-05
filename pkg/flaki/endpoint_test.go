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
	"github.com/stretchr/testify/assert"
)

func TestNextIDEndpoint(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockComponent = mock.NewComponent(mockCtrl)

	var e = MakeNextIDEndpoint(mockComponent)

	rand.Seed(time.Now().UnixNano())
	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), CorrelationIDKey, corrID)
	var req = createFlakiRequest()

	// NextID.
	{
		mockComponent.EXPECT().NextID(ctx, req).Return(createFlakiReply(flakiID), nil).Times(1)
		var reply, err = e(ctx, req)
		assert.Nil(t, err)
		var r = reply.(*fb.FlakiReply)
		assert.Equal(t, flakiID, string(r.Id()))
	}

	// NextID error.
	{
		mockComponent.EXPECT().NextID(ctx, req).Return(nil, fmt.Errorf("fail")).Times(1)
		var reply, err = e(ctx, req)
		assert.NotNil(t, err)
		assert.Nil(t, reply)
	}

	// Wrong request type.
	{
		var reply, err = e(ctx, nil)
		assert.NotNil(t, err)
		assert.Nil(t, reply)
	}
}

func TestNextValidIDEndpoint(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockComponent = mock.NewComponent(mockCtrl)

	var e = MakeNextValidIDEndpoint(mockComponent)

	rand.Seed(time.Now().UnixNano())
	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), CorrelationIDKey, corrID)
	var req = createFlakiRequest()

	// NextValidID.
	mockComponent.EXPECT().NextValidID(ctx, req).Return(createFlakiReply(flakiID)).Times(1)
	var reply, err = e(ctx, req)
	assert.Nil(t, err)
	var r = reply.(*fb.FlakiReply)
	assert.Equal(t, flakiID, string(r.Id()))

	// Wrong request type.
	reply, err = e(ctx, nil)
	assert.NotNil(t, err)
	assert.Nil(t, reply)
}
