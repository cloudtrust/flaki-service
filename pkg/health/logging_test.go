package health_test

//go:generate mockgen -destination=./mock/logging.go -package=mock -mock_names=Logger=Logger github.com/go-kit/kit/log Logger

import (
	"context"
	"math/rand"
	"strconv"
	"testing"
	"time"

	. "github.com/cloudtrust/flaki-service/pkg/health"
	"github.com/cloudtrust/flaki-service/pkg/health/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	common "github.com/cloudtrust/common-healthcheck"
)

func TestEndpointLoggingMW(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockLogger = mock.NewLogger(mockCtrl)
	var mockComponent = mock.NewHealthChecker(mockCtrl)

	var m = MakeEndpointLoggingMW(mockLogger)(MakeExecInfluxHealthCheckEndpoint(mockComponent))

	// Context with correlation ID.
	rand.Seed(time.Now().UnixNano())
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", corrID)
	var rep = []Report{{Name: "influx", Duration: (1 * time.Second).String(), Status: OK.String()}}

	// With correlation ID.
	mockLogger.EXPECT().Log("correlation_id", corrID, "took", gomock.Any()).Return(nil).Times(1)
	mockComponent.EXPECT().ExecInfluxHealthChecks(ctx).Return(rep).Times(1)
	m(ctx, nil)

	// Without correlation ID.
	mockComponent.EXPECT().ExecInfluxHealthChecks(context.Background()).Return(rep).Times(1)
	var f = func() {
		m(context.Background(), nil)
	}
	assert.Panics(t, f)
}

func TestComponentLoggingMW(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockLogger = mock.NewLogger(mockCtrl)
	var mockComponent = mock.NewHealthChecker(mockCtrl)

	var m = MakeComponentLoggingMW(mockLogger)(mockComponent)

	rand.Seed(time.Now().UnixNano())
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", corrID)
	var rep = func(name string) []Report {
		return []Report{{Name: name, Duration: (1 * time.Second).String(), Status: OK.String()}}
	}

	// InfluxHealthChecks.
	{
		mockComponent.EXPECT().ExecInfluxHealthChecks(ctx).Return(rep("influx")).Times(1)
		mockLogger.EXPECT().Log("unit", "ExecInfluxHealthChecks", "correlation_id", corrID, "took", gomock.Any()).Return(nil).Times(1)
		m.ExecInfluxHealthChecks(ctx)

		// Without correlation ID.
		mockComponent.EXPECT().ExecInfluxHealthChecks(context.Background()).Return(rep("influx")).Times(1)
		var f = func() {
			m.ExecInfluxHealthChecks(context.Background())
		}
		assert.Panics(t, f)
	}

	// JaegerHealthChecks.
	{
		mockComponent.EXPECT().ExecJaegerHealthChecks(ctx).Return(rep("jaeger")).Times(1)
		mockLogger.EXPECT().Log("unit", "ExecJaegerHealthChecks", "correlation_id", corrID, "took", gomock.Any()).Return(nil).Times(1)
		m.ExecJaegerHealthChecks(ctx)

		// Without correlation ID.
		mockComponent.EXPECT().ExecJaegerHealthChecks(context.Background()).Return(rep("jaeger")).Times(1)
		var f = func() {
			m.ExecJaegerHealthChecks(context.Background())
		}
		assert.Panics(t, f)
	}

	// RedisHealthChecks.
	{
		mockComponent.EXPECT().ExecRedisHealthChecks(ctx).Return(rep("redis")).Times(1)
		mockLogger.EXPECT().Log("unit", "ExecRedisHealthChecks", "correlation_id", corrID, "took", gomock.Any()).Return(nil).Times(1)
		m.ExecRedisHealthChecks(ctx)

		// Without correlation ID.
		mockComponent.EXPECT().ExecRedisHealthChecks(context.Background()).Return(rep("redis")).Times(1)
		var f = func() {
			m.ExecRedisHealthChecks(context.Background())
		}
		assert.Panics(t, f)
	}

	// SentryHealthChecks.
	{
		mockComponent.EXPECT().ExecSentryHealthChecks(ctx).Return(rep("sentry")).Times(1)
		mockLogger.EXPECT().Log("unit", "ExecSentryHealthChecks", "correlation_id", corrID, "took", gomock.Any()).Return(nil).Times(1)
		m.ExecSentryHealthChecks(ctx)

		// Without correlation ID.
		mockComponent.EXPECT().ExecSentryHealthChecks(context.Background()).Return(rep("sentry")).Times(1)
		var f = func() {
			m.ExecSentryHealthChecks(context.Background())
		}
		assert.Panics(t, f)
	}

	// AllHealthChecks.
	{
		var reply = map[string]string{"influx": "OK", "jaeger": "OK", "redis": "OK", "sentry": "OK"}
		mockComponent.EXPECT().AllHealthChecks(ctx).Return(reply).Times(1)
		mockLogger.EXPECT().Log("unit", "AllHealthChecks", "correlation_id", corrID, "took", gomock.Any()).Return(nil).Times(1)
		m.AllHealthChecks(ctx)

		// Without correlation ID.
		mockComponent.EXPECT().AllHealthChecks(context.Background()).Return(reply).Times(1)
		var f = func() {
			m.AllHealthChecks(context.Background())
		}
		assert.Panics(t, f)
	}
}

func TestInfluxModuleLoggingMW(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockLogger = mock.NewLogger(mockCtrl)
	var mockModule = mock.NewInfluxHealthChecker(mockCtrl)

	var m = MakeInfluxModuleLoggingMW(mockLogger)(mockModule)

	// Context with correlation ID.
	rand.Seed(time.Now().UnixNano())
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", corrID)
	var rep = []common.InfluxReport{{Name: "influx", Duration: (1 * time.Second), Status: common.OK}}

	mockModule.EXPECT().HealthChecks(ctx).Return(rep).Times(1)
	mockLogger.EXPECT().Log("unit", "HealthChecks", "correlation_id", corrID, "took", gomock.Any()).Return(nil).Times(1)
	m.HealthChecks(ctx)

	// Without correlation ID.
	mockModule.EXPECT().HealthChecks(context.Background()).Return(rep).Times(1)
	var f = func() {
		m.HealthChecks(context.Background())
	}
	assert.Panics(t, f)
}

func TestJaegerModuleLoggingMW(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockLogger = mock.NewLogger(mockCtrl)
	var mockModule = mock.NewJaegerHealthChecker(mockCtrl)

	var m = MakeJaegerModuleLoggingMW(mockLogger)(mockModule)

	// Context with correlation ID.
	rand.Seed(time.Now().UnixNano())
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", corrID)
	var rep = []common.JaegerReport{{Name: "jaeger", Duration: (1 * time.Second), Status: common.OK}}

	mockModule.EXPECT().HealthChecks(ctx).Return(rep).Times(1)
	mockLogger.EXPECT().Log("unit", "HealthChecks", "correlation_id", corrID, "took", gomock.Any()).Return(nil).Times(1)
	m.HealthChecks(ctx)

	// Without correlation ID.
	mockModule.EXPECT().HealthChecks(context.Background()).Return(rep).Times(1)
	var f = func() {
		m.HealthChecks(context.Background())
	}
	assert.Panics(t, f)
}

func TestRedisModuleLoggingMW(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockLogger = mock.NewLogger(mockCtrl)
	var mockModule = mock.NewRedisHealthChecker(mockCtrl)

	var m = MakeRedisModuleLoggingMW(mockLogger)(mockModule)

	// Context with correlation ID.
	rand.Seed(time.Now().UnixNano())
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", corrID)
	var rep = []common.RedisReport{{Name: "redis", Duration: (1 * time.Second), Status: common.OK}}

	mockModule.EXPECT().HealthChecks(ctx).Return(rep).Times(1)
	mockLogger.EXPECT().Log("unit", "HealthChecks", "correlation_id", corrID, "took", gomock.Any()).Return(nil).Times(1)
	m.HealthChecks(ctx)

	// Without correlation ID.
	mockModule.EXPECT().HealthChecks(context.Background()).Return(rep).Times(1)
	var f = func() {
		m.HealthChecks(context.Background())
	}
	assert.Panics(t, f)
}

func TestSentryModuleLoggingMW(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockLogger = mock.NewLogger(mockCtrl)
	var mockModule = mock.NewSentryHealthChecker(mockCtrl)

	var m = MakeSentryModuleLoggingMW(mockLogger)(mockModule)

	// Context with correlation ID.
	rand.Seed(time.Now().UnixNano())
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", corrID)
	var rep = []common.SentryReport{{Name: "sentry", Duration: (1 * time.Second), Status: common.OK}}

	mockModule.EXPECT().HealthChecks(ctx).Return(rep).Times(1)
	mockLogger.EXPECT().Log("unit", "HealthChecks", "correlation_id", corrID, "took", gomock.Any()).Return(nil).Times(1)
	m.HealthChecks(ctx)

	// Without correlation ID.
	mockModule.EXPECT().HealthChecks(context.Background()).Return(rep).Times(1)
	var f = func() {
		m.HealthChecks(context.Background())
	}
	assert.Panics(t, f)
}
