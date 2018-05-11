package health_test

//go:generate mockgen -destination=./mock/module.go -package=mock -mock_names=InfluxHealthChecker=InfluxHealthChecker,JaegerHealthChecker=JaegerHealthChecker,RedisHealthChecker=RedisHealthChecker,SentryHealthChecker=SentryHealthChecker,StorageModule=StorageModule  github.com/cloudtrust/flaki-service/pkg/health InfluxHealthChecker,JaegerHealthChecker,RedisHealthChecker,SentryHealthChecker,StorageModule


import (
	"encoding/json"
	"context"
	"testing"
	"time"

	. "github.com/cloudtrust/flaki-service/pkg/health"
	"github.com/cloudtrust/flaki-service/pkg/health/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	common "github.com/cloudtrust/common-healthcheck"
)

func TestHealthChecks(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockInfluxModule = mock.NewInfluxHealthChecker(mockCtrl)
	var mockJaegerModule = mock.NewJaegerHealthChecker(mockCtrl)
	var mockRedisModule = mock.NewRedisHealthChecker(mockCtrl)
	var mockSentryModule = mock.NewSentryHealthChecker(mockCtrl)
	var mockStorage = mock.NewStorageModule(mockCtrl)
	var m = map[string]time.Duration{
		"influx": 1 * time.Minute,
		"jaeger": 1 * time.Minute,
		"redis":  1 * time.Minute,
		"sentry": 1 * time.Minute,
	}

	var c = NewComponent(mockInfluxModule, mockJaegerModule, mockRedisModule, mockSentryModule, mockStorage, m)

	var ( 
		influxReports    = []common.InfluxReport{{Name: "influx", Duration: time.Duration(1 * time.Second), Status: common.OK}}
		jaegerReports    = []common.JaegerReport{{Name: "jaeger", Duration: time.Duration(1 * time.Second), Status: common.OK}}
		redisReports     = []common.RedisReport{{Name: "redis", Duration: time.Duration(1 * time.Second), Status: common.OK}}
		sentryReports    = []common.SentryReport{{Name: "sentry", Duration: time.Duration(1 * time.Second), Status: common.OK}}
		makeStoredReport = func(name string) StoredReport {
			return StoredReport{
				ComponentID: "000-000-000-00",
				ComponentName: "flaki",
				HealthcheckUnit: name,
				Reports: json.RawMessage(`{}`),
				LastUpdated: time.Now(),
				ValidUntil: time.Now().Add(1 * time.Hour),
			}
		}
	)

	// Influx.
	mockInfluxModule.EXPECT().HealthChecks(context.Background()).Return(influxReports).Times(1)
	mockStorage.EXPECT().Update("influx", m["influx"], gomock.Any()).Times(1)
	{
		var report = c.ExecInfluxHealthChecks(context.Background())
		var json, _ = json.Marshal(&report)
		assert.Equal(t, `[{"Name":"influx","Duration":1000000000,"Status":0,"Error":null}]`, string(json))
	}

	// Jaeger.
	mockJaegerModule.EXPECT().HealthChecks(context.Background()).Return(jaegerReports).Times(1)
	mockStorage.EXPECT().Update("jaeger", m["jaeger"], gomock.Any()).Times(1)
	{
		var report = c.ExecJaegerHealthChecks(context.Background())
		var json, _ = json.Marshal(&report)
		assert.Equal(t, `[{"Name":"jaeger","Duration":1000000000,"Status":0,"Error":null}]`, string(json))
	}

	// Redis.
	mockRedisModule.EXPECT().HealthChecks(context.Background()).Return(redisReports).Times(1)
	mockStorage.EXPECT().Update("redis", m["redis"], gomock.Any()).Times(1)
	{
		var report = c.ExecRedisHealthChecks(context.Background())
		var json, _ = json.Marshal(&report)
		assert.Equal(t, `[{"Name":"redis","Duration":1000000000,"Status":0,"Error":null}]`, string(json))

	}

	// Sentry.
	mockSentryModule.EXPECT().HealthChecks(context.Background()).Return(sentryReports).Times(1)
	mockStorage.EXPECT().Update("sentry", m["sentry"], gomock.Any()).Times(1)
	{
		var report = c.ExecSentryHealthChecks(context.Background())
		var json, _ = json.Marshal(&report)
		assert.Equal(t, `[{"Name":"sentry","Duration":1000000000,"Status":0,"Error":null}]`, string(json))
	}

	// All.
	mockStorage.EXPECT().Read("influx").Return(makeStoredReport("influx"), nil).Times(1)
	mockStorage.EXPECT().Read("jaeger").Return(makeStoredReport("jaeger"), nil).Times(1)
	mockStorage.EXPECT().Read("redis").Return(makeStoredReport("redis"), nil).Times(1)
	mockStorage.EXPECT().Read("sentry").Return(makeStoredReport("sentry"), nil).Times(1)
	{
		var reply = c.AllHealthChecks(context.Background())
		assert.Equal(t, "OK", reply["influx"])
		assert.Equal(t, "OK", reply["jaeger"])
		assert.Equal(t, "OK", reply["redis"])
		assert.Equal(t, "OK", reply["sentry"])
	}
}

// func TestHealthChecksFail(t *testing.T) {
// 	var mockCtrl = gomock.NewController(t)
// 	defer mockCtrl.Finish()
// 	var mockInfluxModule = mock.NewInfluxHealthChecker(mockCtrl)
// 	var mockJaegerModule = mock.NewJaegerHealthChecker(mockCtrl)
// 	var mockRedisModule = mock.NewRedisHealthChecker(mockCtrl)
// 	var mockSentryModule = mock.NewSentryHealthChecker(mockCtrl)
// 	var mockStorage = mock.NewStorageModule(mockCtrl)
// 	var m = map[string]time.Duration{
// 		"influx": 1 * time.Minute,
// 		"jaeger": 1 * time.Minute,
// 		"redis":  1 * time.Minute,
// 		"sentry": 1 * time.Minute,
// 	}

// 	var c = NewComponent(mockInfluxModule, mockJaegerModule, mockRedisModule, mockSentryModule, mockStorage, m)

// 	var (
// 		influxReports    = []common.InfluxReport{{Name: "influx", Duration: time.Duration(1 * time.Second), Status: common.Deactivated}}
// 		jaegerReports    = []common.JaegerReport{{Name: "jaeger", Duration: time.Duration(1 * time.Second), Status: common.KO, Error: fmt.Errorf("fail")}}
// 		redisReports     = []common.RedisReport{{Name: "redis", Duration: time.Duration(1 * time.Second), Status: common.Degraded, Error: fmt.Errorf("fail")}}
// 		sentryReports    = []common.SentryReport{{Name: "sentry", Duration: time.Duration(1 * time.Second), Status: common.KO, Error: fmt.Errorf("fail")}}
// 		makeStoredReport = func(name string) StoredReport {
// 			return StoredReport{
// 				ComponentID: "000-000-000-00",
// 				ComponentName: "flaki",
// 				HealthcheckUnit: name,
// 				Reports: json.RawMessage(`{}`),
// 				LastUpdated: time.Now(),
// 				ValidUntil: time.Now().Add(1 * time.Hour),
// 			}
// 		}
// 	)

// 	// Influx.
// 	mockInfluxModule.EXPECT().HealthChecks(context.Background()).Return(influxReports).Times(1)
// 	mockStorage.EXPECT().Update("influx", m["influx"], gomock.Any()).Times(1)
// 	{
// 		var report = c.ExecInfluxHealthChecks(context.Background())
// 		var json, _ = json.Marshal(&report)
// 		assert.Equal(t, `[{"Name":"influx","Duration":1000000000,"Status":3,"Error":null}]`, string(json))
// 	}

// 	// Jaeger.
// 	mockJaegerModule.EXPECT().HealthChecks(context.Background()).Return(jaegerReports).Times(1)
// 	mockStorage.EXPECT().Update("jaeger", m["jaeger"], gomock.Any()).Times(1)
// 	{
// 		var report = c.ExecJaegerHealthChecks(context.Background())
// 		var json, _ = json.Marshal(&report)
// 		assert.Equal(t, `[{"Name":"jaeger","Duration":1000000000,"Status":1,"Error":TODO]`, string(json))
// 	}

// 	// Redis.
// 	mockRedisModule.EXPECT().HealthChecks(context.Background()).Return(redisReports).Times(1)
// 	mockStorage.EXPECT().Update("redis", m["redis"], gomock.Any()).Times(1)
// 	{
// 		var report = c.ExecRedisHealthChecks(context.Background())
// 		var json, _ = json.Marshal(&report)
// 		assert.Equal(t, `[{"Name":"redis","Duration":1000000000,"Status":1,"Error":TODO]`, string(json))
// 	}

// 	// Sentry.
// 	mockSentryModule.EXPECT().HealthChecks(context.Background()).Return(sentryReports).Times(1)
// 	mockStorage.EXPECT().Update("sentry", m["sentry"], gomock.Any()).Times(1)
// 	{
// 		var report = c.ExecSentryHealthChecks(context.Background())
// 		var json, _ = json.Marshal(&report)
// 		assert.Equal(t, `[{"Name":"sentry","Duration":1000000000,"Status":1,"Error":TODO]`, string(json))
// 	}

// 	// All.
// 	mockStorage.EXPECT().Read("influx").Return(makeStoredReport("influx"), nil).Times(1)
// 	mockStorage.EXPECT().Read("jaeger").Return(makeStoredReport("jaeger"), nil).Times(1)
// 	mockStorage.EXPECT().Read("redis").Return(makeStoredReport("redis"), nil).Times(1)
// 	mockStorage.EXPECT().Read("sentry").Return(makeStoredReport("sentry"), nil).Times(1)
// 	{
// 		var reply = c.AllHealthChecks(context.Background())
// 		assert.Equal(t, "Deactivated", reply["influx"])
// 		assert.Equal(t, "KO", reply["jaeger"])
// 		assert.Equal(t, "Degraded", reply["redis"])
// 		assert.Equal(t, "KO", reply["sentry"])
// 	}
// }
