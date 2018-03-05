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
	var ir = c.InfluxHealthChecks(context.Background()).Reports[0]
	assert.Equal(t, "influx", ir.Name)
	assert.NotZero(t, ir.Duration)
	assert.Equal(t, OK, ir.Status)
	assert.Zero(t, ir.Error)

	// Jaeger.
	var jr = c.JaegerHealthChecks(context.Background()).Reports[0]
	assert.Equal(t, "jaeger", jr.Name)
	assert.NotZero(t, jr.Duration)
	assert.Equal(t, OK, jr.Status)
	assert.Zero(t, jr.Error)

	// Redis.
	var rr = c.RedisHealthChecks(context.Background()).Reports[0]
	assert.Equal(t, "redis", rr.Name)
	assert.NotZero(t, rr.Duration)
	assert.Equal(t, OK, rr.Status)
	assert.Zero(t, rr.Error)

	// Sentry.
	var sr = c.SentryHealthChecks(context.Background()).Reports[0]
	assert.Equal(t, "sentry", sr.Name)
	assert.NotZero(t, sr.Duration)
	assert.Equal(t, OK, sr.Status)
	assert.Zero(t, sr.Error)
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
	var ir = c.InfluxHealthChecks(context.Background()).Reports[0]
	assert.Equal(t, "influx", ir.Name)
	assert.NotZero(t, ir.Duration)
	assert.Equal(t, KO, ir.Status)
	assert.Equal(t, "fail", ir.Error)

	// Jaeger.
	var jr = c.JaegerHealthChecks(context.Background()).Reports[0]
	assert.Equal(t, "jaeger", jr.Name)
	assert.NotZero(t, jr.Duration)
	assert.Equal(t, KO, jr.Status)
	assert.Equal(t, "fail", jr.Error)

	// Redis.
	var rr = c.RedisHealthChecks(context.Background()).Reports[0]
	assert.Equal(t, "redis", rr.Name)
	assert.NotZero(t, rr.Duration)
	assert.Equal(t, KO, rr.Status)
	assert.Equal(t, "fail", rr.Error)

	// Sentry.
	var sr = c.SentryHealthChecks(context.Background()).Reports[0]
	assert.Equal(t, "sentry", sr.Name)
	assert.NotZero(t, sr.Duration)
	assert.Equal(t, KO, sr.Status)
	assert.Equal(t, "fail", sr.Error)
}
