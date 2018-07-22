package health_test

//go:generate mockgen -destination=./mock/idGenerator.go -package=mock -mock_names=IDGeneratorModule=FlakiModule github.com/cloudtrust/flaki-service/pkg/flaki IDGeneratorModule

import (
	"context"
	"encoding/json"
	"math/rand"
	"strconv"
	"testing"
	"time"

	. "github.com/cloudtrust/flaki-service/pkg/health"
	"github.com/cloudtrust/flaki-service/pkg/health/mock"
	"github.com/golang/mock/gomock"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestEndpointCorrelationIDMW(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockComponent = mock.NewHealthCheckers(mockCtrl)
	var mockFlakiModule = mock.NewFlakiModule(mockCtrl)

	var m = MakeEndpointCorrelationIDMW(mockFlakiModule)(MakeHealthChecksEndpoint(mockComponent))

	var (
		flakiID = strconv.FormatUint(rand.Uint64(), 10)
		corrID  = strconv.FormatUint(rand.Uint64(), 10)
		ctx     = context.WithValue(context.Background(), "correlation_id", corrID)
		ctxFID  = context.WithValue(context.Background(), "correlation_id", flakiID)

		req = map[string]string{
			"module": "cockroach",
		}
		rep = json.RawMessage(`{"key":"value"}`)
	)

	// Context with correlation ID.
	mockComponent.EXPECT().HealthChecks(ctx, req).Return(rep, nil).Times(1)
	m(ctx, req)

	// Without correlation ID.
	mockFlakiModule.EXPECT().NextValidID(gomock.Any()).Return(flakiID).Times(1)
	mockComponent.EXPECT().HealthChecks(ctxFID, req).Return(rep, nil).Times(1)
	m(context.Background(), req)
}
