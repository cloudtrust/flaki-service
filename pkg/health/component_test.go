package health_test

import (
	"context"
	"fmt"
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
	var mockInfluxModule = mock.NewInfluxHealthChecker(mockCtrl)
	var mockJaegerModule = mock.NewJaegerHealthChecker(mockCtrl)
	var mockRedisModule = mock.NewRedisHealthChecker(mockCtrl)
	var mockSentryModule = mock.NewSentryHealthChecker(mockCtrl)
	var mockStorage = mock.NewStorageModule(mockCtrl)

	var c = NewComponent(mockInfluxModule, mockJaegerModule, mockRedisModule, mockSentryModule, mockStorage)

	var (
		influxReports    = []InfluxReport{{Name: "influx", Duration: time.Duration(1 * time.Second), Status: OK}}
		jaegerReports    = []JaegerReport{{Name: "jaeger", Duration: time.Duration(1 * time.Second), Status: OK}}
		redisReports     = []RedisReport{{Name: "redis", Duration: time.Duration(1 * time.Second), Status: OK}}
		sentryReports    = []SentryReport{{Name: "sentry", Duration: time.Duration(1 * time.Second), Status: OK}}
		makeStoredReport = func(name string, s Status) []StoredReport {
			return []StoredReport{{Name: name, Duration: 1 * time.Second, Status: s, Error: "", LastExecution: time.Now(), ValidUntil: time.Now().Add(1 * time.Hour)}}
		}
	)

	// Influx.
	mockInfluxModule.EXPECT().HealthChecks(context.Background()).Return(influxReports).Times(1)
	mockStorage.EXPECT().Update("influx", gomock.Any()).Times(1)
	mockStorage.EXPECT().Clean().Times(1)
	{
		var report = c.ExecInfluxHealthChecks(context.Background())[0]
		assert.Equal(t, "influx", report.Name)
		assert.NotZero(t, report.Duration)
		assert.Equal(t, "OK", report.Status)
		assert.Zero(t, report.Error)
	}

	// Jaeger.
	mockJaegerModule.EXPECT().HealthChecks(context.Background()).Return(jaegerReports).Times(1)
	mockStorage.EXPECT().Update("jaeger", gomock.Any()).Times(1)
	mockStorage.EXPECT().Clean().Times(1)
	{
		var report = c.ExecJaegerHealthChecks(context.Background())[0]
		assert.Equal(t, "jaeger", report.Name)
		assert.NotZero(t, report.Duration)
		assert.Equal(t, "OK", report.Status)
		assert.Zero(t, report.Error)
	}

	// Redis.
	mockRedisModule.EXPECT().HealthChecks(context.Background()).Return(redisReports).Times(1)
	mockStorage.EXPECT().Update("redis", gomock.Any()).Times(1)
	mockStorage.EXPECT().Clean().Times(1)
	{
		var report = c.ExecRedisHealthChecks(context.Background())[0]
		assert.Equal(t, "redis", report.Name)
		assert.NotZero(t, report.Duration)
		assert.Equal(t, "OK", report.Status)
		assert.Zero(t, report.Error)
	}

	// Sentry.
	mockSentryModule.EXPECT().HealthChecks(context.Background()).Return(sentryReports).Times(1)
	mockStorage.EXPECT().Update("sentry", gomock.Any()).Times(1)
	mockStorage.EXPECT().Clean().Times(1)
	{
		var report = c.ExecSentryHealthChecks(context.Background())[0]
		assert.Equal(t, "sentry", report.Name)
		assert.NotZero(t, report.Duration)
		assert.Equal(t, "OK", report.Status)
		assert.Zero(t, report.Error)
	}

	// All.
	mockStorage.EXPECT().Read("influx").Return(makeStoredReport("influx", OK), nil).Times(1)
	mockStorage.EXPECT().Read("jaeger").Return(makeStoredReport("jaeger", OK), nil).Times(1)
	mockStorage.EXPECT().Read("redis").Return(makeStoredReport("redis", OK), nil).Times(1)
	mockStorage.EXPECT().Read("sentry").Return(makeStoredReport("sentry", OK), nil).Times(1)
	{
		var reply = c.AllHealthChecks(context.Background())
		assert.Equal(t, "OK", reply["influx"])
		assert.Equal(t, "OK", reply["jaeger"])
		assert.Equal(t, "OK", reply["redis"])
		assert.Equal(t, "OK", reply["sentry"])
	}
}

func TestHealthChecksFail(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockInfluxModule = mock.NewInfluxHealthChecker(mockCtrl)
	var mockJaegerModule = mock.NewJaegerHealthChecker(mockCtrl)
	var mockRedisModule = mock.NewRedisHealthChecker(mockCtrl)
	var mockSentryModule = mock.NewSentryHealthChecker(mockCtrl)
	var mockStorage = mock.NewStorageModule(mockCtrl)

	var c = NewComponent(mockInfluxModule, mockJaegerModule, mockRedisModule, mockSentryModule, mockStorage)

	var (
		influxReports    = []InfluxReport{{Name: "influx", Duration: time.Duration(1 * time.Second), Status: Deactivated}}
		jaegerReports    = []JaegerReport{{Name: "jaeger", Duration: time.Duration(1 * time.Second), Status: KO, Error: fmt.Errorf("fail")}}
		redisReports     = []RedisReport{{Name: "redis", Duration: time.Duration(1 * time.Second), Status: Degraded, Error: fmt.Errorf("fail")}}
		sentryReports    = []SentryReport{{Name: "sentry", Duration: time.Duration(1 * time.Second), Status: KO, Error: fmt.Errorf("fail")}}
		makeStoredReport = func(name string, s Status) []StoredReport {
			return []StoredReport{{Name: name, Duration: 1 * time.Second, Status: s, Error: "fail", LastExecution: time.Now(), ValidUntil: time.Now().Add(1 * time.Hour)}}
		}
	)

	// Influx.
	mockInfluxModule.EXPECT().HealthChecks(context.Background()).Return(influxReports).Times(1)
	mockStorage.EXPECT().Update("influx", gomock.Any()).Times(1)
	mockStorage.EXPECT().Clean().Times(1)
	{
		var report = c.ExecInfluxHealthChecks(context.Background())[0]
		assert.Equal(t, "influx", report.Name)
		assert.NotZero(t, report.Duration)
		assert.Equal(t, "Deactivated", report.Status)
		assert.Zero(t, report.Error)
	}

	// Jaeger.
	mockJaegerModule.EXPECT().HealthChecks(context.Background()).Return(jaegerReports).Times(1)
	mockStorage.EXPECT().Update("jaeger", gomock.Any()).Times(1)
	mockStorage.EXPECT().Clean().Times(1)
	{
		var report = c.ExecJaegerHealthChecks(context.Background())[0]
		assert.Equal(t, "jaeger", report.Name)
		assert.NotZero(t, report.Duration)
		assert.Equal(t, "KO", report.Status)
		assert.Equal(t, "fail", report.Error)
	}

	// Redis.
	mockRedisModule.EXPECT().HealthChecks(context.Background()).Return(redisReports).Times(1)
	mockStorage.EXPECT().Update("redis", gomock.Any()).Times(1)
	mockStorage.EXPECT().Clean().Times(1)
	{
		var report = c.ExecRedisHealthChecks(context.Background())[0]
		assert.Equal(t, "redis", report.Name)
		assert.NotZero(t, report.Duration)
		assert.Equal(t, "Degraded", report.Status)
		assert.Equal(t, "fail", report.Error)
	}

	// Sentry.
	mockSentryModule.EXPECT().HealthChecks(context.Background()).Return(sentryReports).Times(1)
	mockStorage.EXPECT().Update("sentry", gomock.Any()).Times(1)
	mockStorage.EXPECT().Clean().Times(1)
	{
		var report = c.ExecSentryHealthChecks(context.Background())[0]
		assert.Equal(t, "sentry", report.Name)
		assert.NotZero(t, report.Duration)
		assert.Equal(t, "KO", report.Status)
		assert.Equal(t, "fail", report.Error)
	}

	// All.
	mockStorage.EXPECT().Read("influx").Return(makeStoredReport("influx", Deactivated), nil).Times(1)
	mockStorage.EXPECT().Read("jaeger").Return(makeStoredReport("jaeger", KO), nil).Times(1)
	mockStorage.EXPECT().Read("redis").Return(makeStoredReport("redis", Degraded), nil).Times(1)
	mockStorage.EXPECT().Read("sentry").Return(makeStoredReport("sentry", KO), nil).Times(1)
	{
		var reply = c.AllHealthChecks(context.Background())
		assert.Equal(t, "Deactivated", reply["influx"])
		assert.Equal(t, "KO", reply["jaeger"])
		assert.Equal(t, "Degraded", reply["redis"])
		assert.Equal(t, "KO", reply["sentry"])
	}
}
