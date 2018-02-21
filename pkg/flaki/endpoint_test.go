package flaki

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/cloudtrust/flaki-service/pkg/flaki/mock"
	"github.com/go-kit/kit/endpoint"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNextIDEndpoint(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockComponent = mock.NewComponent(mockCtrl)

	rand.Seed(time.Now().UnixNano())
	var flakiID = strconv.FormatUint(rand.Uint64(), 10)

	// Context with correlation ID.
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", corrID)

	var e = MakeNextIDEndpoint(mockComponent)

	// NextID.
	mockComponent.EXPECT().NextID(ctx).Return(flakiID, nil).Times(1)

	var id, err = e(ctx, nil)
	assert.Nil(t, err)
	assert.Equal(t, flakiID, id)
}

func TestNextValidIDEndpoint(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockComponent = mock.NewComponent(mockCtrl)

	rand.Seed(time.Now().UnixNano())
	var flakiID = strconv.FormatUint(rand.Uint64(), 10)

	// Context with correlation ID.
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", corrID)

	var e = MakeNextValidIDEndpoint(mockComponent)

	// NextValidID.
	mockComponent.EXPECT().NextValidID(ctx).Return(flakiID).Times(1)

	var id, err = e(ctx, nil)
	assert.Nil(t, err)
	assert.Equal(t, flakiID, id)
}

func MakeMockEndpoint(id string, fail bool) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		if fail {
			return "", fmt.Errorf("fail")
		}
		return id, nil
	}
}
