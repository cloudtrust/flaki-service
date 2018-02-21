package health_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/cloudtrust/flaki-service/pkg/health"
	"github.com/cloudtrust/flaki-service/pkg/health/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestSentryHealthChecks(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockSentry = mock.NewSentry(mockCtrl)

	mockSentry.EXPECT().URL().Return("https://a:b@sentry.io/api/1/store/").Times(1)

	var s = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer s.Close()

	var m = NewSentryModule(mockSentry, s.Client())

	var report = m.HealthChecks(context.Background())[0]
	assert.Equal(t, "ping", report.Name)
	assert.NotZero(t, report.Duration)
	assert.Equal(t, OK, report.Status)
	assert.Zero(t, report.Error)
}

func TestNoopSentryHealthChecks(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockSentry = mock.NewSentry(mockCtrl)

	mockSentry.EXPECT().URL().Return("").Times(1)

	var s = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer s.Close()

	var m = NewSentryModule(mockSentry, s.Client())

	var report = m.HealthChecks(context.Background())[0]
	assert.Equal(t, "ping", report.Name)
	assert.Equal(t, "N/A", report.Duration)
	assert.Equal(t, Deactivated, report.Status)
	assert.Zero(t, report.Error)
}
