package health_test

import (
	"context"
	"math/rand"
	"strconv"
	"testing"
	"time"

	. "github.com/cloudtrust/flaki-service/pkg/health"
	"github.com/cloudtrust/flaki-service/pkg/health/mock"
	"github.com/golang/mock/gomock"
)

func TestEndpointCorrelationIDMW(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockComponent = mock.NewComponent(mockCtrl)
	var mockFlakiModule = mock.NewFlakiModule(mockCtrl)

	var m = MakeEndpointCorrelationIDMW(mockFlakiModule)(MakeInfluxHealthCheckEndpoint(mockComponent))

	rand.Seed(time.Now().UnixNano())
	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", corrID)
	var ctxFID = context.WithValue(context.Background(), "correlation_id", flakiID)

	// Context with correlation ID.
	mockComponent.EXPECT().InfluxHealthChecks(ctx).Return(Reports{}).Times(1)
	m(ctx, nil)

	// Without correlation ID.
	mockFlakiModule.EXPECT().NextValidID(gomock.Any()).Return(flakiID).Times(1)
	mockComponent.EXPECT().InfluxHealthChecks(ctxFID).Return(Reports{}).Times(1)
	m(context.Background(), nil)
}
