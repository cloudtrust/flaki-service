package health_test

import (
	"context"
	"testing"
	"time"

	. "github.com/cloudtrust/flaki-service/pkg/health"
	"github.com/cloudtrust/flaki-service/pkg/health/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestHealthChecks(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockInfluxModule = mock.NewInfluxModule(mockCtrl)
	var mockJaegerModule = mock.NewJaegerModule(mockCtrl)
	var mockRedisModule = mock.NewRedisModule(mockCtrl)
	var mockSentryModule = mock.NewSentryModule(mockCtrl)

	mockInfluxModule.EXPECT().HealthChecks(context.Background()).Return([]InfluxReport{{Name: "influx", Duration: time.Duration(1 * time.Second).String(), Status: OK}}).Times(1)
	mockJaegerModule.EXPECT().HealthChecks(context.Background()).Return([]JaegerReport{{Name: "jaeger", Duration: time.Duration(1 * time.Second).String(), Status: OK}}).Times(1)
	mockRedisModule.EXPECT().HealthChecks(context.Background()).Return([]RedisReport{{Name: "redis", Duration: time.Duration(1 * time.Second).String(), Status: OK}}).Times(1)
	mockSentryModule.EXPECT().HealthChecks(context.Background()).Return([]SentryReport{{Name: "sentry", Duration: time.Duration(1 * time.Second).String(), Status: OK}}).Times(1)

	var c = NewComponent(mockInfluxModule, mockJaegerModule, mockRedisModule, mockSentryModule)

	// Influx.
	{
		var report = c.InfluxHealthChecks(context.Background()).Reports[0]
		assert.Equal(t, "influx", report.Name)
		assert.NotZero(t, report.Duration)
		assert.Equal(t, OK, report.Status)
		assert.Zero(t, report.Error)
	}

	// Jaeger.
	{
		var report = c.JaegerHealthChecks(context.Background()).Reports[0]
		assert.Equal(t, "jaeger", report.Name)
		assert.NotZero(t, report.Duration)
		assert.Equal(t, OK, report.Status)
		assert.Zero(t, report.Error)
	}

	// Redis.
	{
		var report = c.RedisHealthChecks(context.Background()).Reports[0]
		assert.Equal(t, "redis", report.Name)
		assert.NotZero(t, report.Duration)
		assert.Equal(t, OK, report.Status)
		assert.Zero(t, report.Error)
	}

	// Sentry.
	{
		var report = c.SentryHealthChecks(context.Background()).Reports[0]
		assert.Equal(t, "sentry", report.Name)
		assert.NotZero(t, report.Duration)
		assert.Equal(t, OK, report.Status)
		assert.Zero(t, report.Error)
	}
}
func TestHealthChecksFail(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockInfluxModule = mock.NewInfluxModule(mockCtrl)
	var mockJaegerModule = mock.NewJaegerModule(mockCtrl)
	var mockRedisModule = mock.NewRedisModule(mockCtrl)
	var mockSentryModule = mock.NewSentryModule(mockCtrl)

	mockInfluxModule.EXPECT().HealthChecks(context.Background()).Return([]InfluxReport{{Name: "influx", Duration: time.Duration(1 * time.Second).String(), Status: KO, Error: "fail"}}).Times(1)
	mockJaegerModule.EXPECT().HealthChecks(context.Background()).Return([]JaegerReport{{Name: "jaeger", Duration: time.Duration(1 * time.Second).String(), Status: KO, Error: "fail"}}).Times(1)
	mockRedisModule.EXPECT().HealthChecks(context.Background()).Return([]RedisReport{{Name: "redis", Duration: time.Duration(1 * time.Second).String(), Status: KO, Error: "fail"}}).Times(1)
	mockSentryModule.EXPECT().HealthChecks(context.Background()).Return([]SentryReport{{Name: "sentry", Duration: time.Duration(1 * time.Second).String(), Status: KO, Error: "fail"}}).Times(1)

	var c = NewComponent(mockInfluxModule, mockJaegerModule, mockRedisModule, mockSentryModule)

	// Influx.
	{
		var report = c.InfluxHealthChecks(context.Background()).Reports[0]
		assert.Equal(t, "influx", report.Name)
		assert.NotZero(t, report.Duration)
		assert.Equal(t, KO, report.Status)
		assert.Equal(t, "fail", report.Error)
	}

	// Jaeger.
	{
		var report = c.JaegerHealthChecks(context.Background()).Reports[0]
		assert.Equal(t, "jaeger", report.Name)
		assert.NotZero(t, report.Duration)
		assert.Equal(t, KO, report.Status)
		assert.Equal(t, "fail", report.Error)
	}

	// Redis.
	{
		var report = c.RedisHealthChecks(context.Background()).Reports[0]
		assert.Equal(t, "redis", report.Name)
		assert.NotZero(t, report.Duration)
		assert.Equal(t, KO, report.Status)
		assert.Equal(t, "fail", report.Error)
	}

	// Sentry.
	{
		var report = c.SentryHealthChecks(context.Background()).Reports[0]
		assert.Equal(t, "sentry", report.Name)
		assert.NotZero(t, report.Duration)
		assert.Equal(t, KO, report.Status)
		assert.Equal(t, "fail", report.Error)
	}
}
