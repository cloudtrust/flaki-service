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

func TestNextID(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockFlaki = mock.NewFlaki(mockCtrl)

	rand.Seed(time.Now().UnixNano())
	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var m = NewModule(mockFlaki)

	mockFlaki.EXPECT().NextIDString().Return(flakiID, nil).Times(1)
	var id, err = m.NextID(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, flakiID, id)
}

func TestNextIDFail(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockFlaki = mock.NewFlaki(mockCtrl)

	var m = NewModule(mockFlaki)

	// When an error is returned, the id is the zero string.
	mockFlaki.EXPECT().NextIDString().Return("", fmt.Errorf("fail")).Times(1)
	var id, err = m.NextID(context.Background())
	assert.NotNil(t, err)
	assert.Zero(t, id)
}

func TestNextValidID(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockFlaki = mock.NewFlaki(mockCtrl)

	rand.Seed(time.Now().UnixNano())
	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var m = NewModule(mockFlaki)

	mockFlaki.EXPECT().NextValidIDString().Return(flakiID).Times(1)
	var id = m.NextValidID(context.Background())
	assert.Equal(t, flakiID, id)
}
