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
)

func TestComponentTrackingMW(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockComponent = mock.NewComponent(mockCtrl)
	var mockSentry = mock.NewSentry(mockCtrl)

	// Context with correlation ID.
	rand.Seed(time.Now().UnixNano())
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", corrID)

	var m = MakeComponentTrackingMW(mockSentry)(mockComponent)

	// NextID.
	mockComponent.EXPECT().NextID(ctx).Return("", fmt.Errorf("fail")).Times(1)
	mockSentry.EXPECT().CaptureError(fmt.Errorf("fail"), map[string]string{"correlation_id": corrID}).Return("").Times(1)
	m.NextID(ctx)

	// NextValidID never returns an error.
	mockComponent.EXPECT().NextValidID(ctx).Return(corrID).Times(1)
	mockSentry.EXPECT().CaptureError(gomock.Any(), gomock.Any()).Times(0)
	m.NextValidID(ctx)

	// NextID without correlation ID.
	mockComponent.EXPECT().NextID(context.Background()).Return("", fmt.Errorf("fail")).Times(1)
	mockSentry.EXPECT().CaptureError(fmt.Errorf("fail"), map[string]string{}).Times(1)
	m.NextID(context.Background())
}
