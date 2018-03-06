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

	var m = MakeComponentTrackingMW(mockSentry)(mockComponent)

	rand.Seed(time.Now().UnixNano())
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), CorrelationIDKey, corrID)
	var req = createFlakiRequest()
	var reply = createFlakiReply(corrID)

	// NextID.
	mockComponent.EXPECT().NextID(ctx, req).Return(nil, fmt.Errorf("fail")).Times(1)
	mockSentry.EXPECT().CaptureError(fmt.Errorf("fail"), map[string]string{TrackingCorrelationIDKey: corrID}).Return("").Times(1)
	m.NextID(ctx, req)

	// NextID without correlation ID.
	mockComponent.EXPECT().NextID(context.Background(), req).Return(nil, fmt.Errorf("fail")).Times(1)
	mockSentry.EXPECT().CaptureError(fmt.Errorf("fail"), map[string]string{}).Times(1)
	m.NextID(context.Background(), req)

	// NextValidID never returns an error.
	mockComponent.EXPECT().NextValidID(ctx, req).Return(reply).Times(1)
	m.NextValidID(ctx, req)
}
