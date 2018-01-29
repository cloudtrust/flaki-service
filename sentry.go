package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	sentry "github.com/getsentry/raven-go"
)

type Sentry interface {
	URL() string
}

func MakeSentryHealthChecks(sentry Sentry) *HealthChecks {

	var checks = []HealthTest{
		makeSentryPingHealthTest(sentry),
	}

	return &HealthChecks{
		name:   "sentry",
		checks: checks,
	}
}

func makeSentryPingHealthTest(sentry Sentry) HealthTest {
	return func() TestReport {

		// Build sentry health url from sentry dsn. The health url is <sentryURL>/_health
		var dsn = sentry.URL()
		var healthURL string
		if idx := strings.LastIndex(dsn, "/api/"); idx != -1 {
			healthURL = fmt.Sprintf("%s/_health", dsn[:idx])
		}

		// Get Sentry health status.
		var now = time.Now()
		var status = getSentryStatus(healthURL)
		var duration = time.Since(now)

		return TestReport{
			Name:     "ping",
			Duration: duration.String(),
			Status:   status,
		}
	}
}

func getSentryStatus(url string) string {
	// Query sentry health endpoint.
	var res *http.Response
	{
		var err error
		res, err = http.DefaultClient.Get(url)
		if err != nil {
			return "KO"
		}
		if res != nil {
			defer res.Body.Close()
		}
	}

	// Chesk response status.
	if res.StatusCode != http.StatusOK {
		return "KO"
	}

	// Chesk response body. The sentry health endpoint returns "ok" when there is no issue.
	var response []byte
	{
		var err error
		response, err = ioutil.ReadAll(res.Body)
		if err != nil {
			return "KO"
		}
	}

	if strings.Compare(string(response), "ok") == 0 {
		return "OK"
	}

	return "KO"
}

// NoopSentry is a sentry client that does nothing.
type NoopSentry struct{}

// CaptureError does nothing for the receiver NoopSentry.
func (s *NoopSentry) CaptureError(err error, tags map[string]string, interfaces ...sentry.Interface) string {
	return ""
}

// CaptureErrorAndWait does nothing for the receiver NoopSentry.
func (s *NoopSentry) CaptureErrorAndWait(err error, tags map[string]string, interfaces ...sentry.Interface) string {
	return ""
}

// URL does nothing for the receiver NoopSentry.
func (s *NoopSentry) URL() string {
	return ""
}

// Close does nothing for the receiver NoopSentry.
func (s *NoopSentry) Close() {}
