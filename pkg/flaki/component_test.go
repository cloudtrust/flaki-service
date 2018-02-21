package flaki

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/cloudtrust/flaki-service/pkg/flaki/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewComponent(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockModule = mock.NewModule(mockCtrl)

	rand.Seed(time.Now().UnixNano())
	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var c = NewComponent(mockModule)

	// NextID.
	mockModule.EXPECT().NextID(context.Background()).Return(flakiID, nil).Times(1)
	var id, err = c.NextID(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, flakiID, id)

	// NextID fail.
	mockModule.EXPECT().NextID(context.Background()).Return("", fmt.Errorf("fail")).Times(1)
	id, err = c.NextID(context.Background())
	assert.NotNil(t, err)
	assert.Zero(t, id)

	// NextValidID.
	mockModule.EXPECT().NextValidID(context.Background()).Return(flakiID).Times(1)
	id = c.NextValidID(context.Background())
	assert.Equal(t, flakiID, id)
}
