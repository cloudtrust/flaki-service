package health

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type SentryHealthReport struct {
	Name     string
	Duration string
	Status   string
	Error    string
}

type Sentry interface {
	URL() string
}

type SentryHealthModule struct {
	sentry Sentry
}

// NewSentryHealthModule returns the sentry health module.
func NewSentryHealthModule(sentry Sentry) *SentryHealthModule {
	return &SentryHealthModule{sentry: sentry}
}

// HealthChecks executes all health checks for Sentry.
func (s *SentryHealthModule) HealthChecks(context.Context) []SentryHealthReport {
	var reports = []SentryHealthReport{}
	reports = append(reports, sentryPingCheck(s.sentry))
	return reports
}

func sentryPingCheck(sentry Sentry) SentryHealthReport {
	// Build sentry health url from sentry dsn. The health url is <sentryURL>/_health
	var dsn = sentry.URL()
	var healthURL string
	if idx := strings.LastIndex(dsn, "/api/"); idx != -1 {
		healthURL = fmt.Sprintf("%s/_health", dsn[:idx])
	}

	// Get Sentry health status.
	var now = time.Now()
	var status, err = getSentryStatus(healthURL)
	var duration = time.Since(now)

	var error = ""
	if err != nil {
		error = err.Error()
	}

	return SentryHealthReport{
		Name:     "ping",
		Duration: duration.String(),
		Status:   status,
		Error:    error,
	}
}

func getSentryStatus(url string) (string, error) {
	// Query sentry health endpoint.
	var res *http.Response
	{
		var err error
		res, err = http.DefaultClient.Get(url)
		if err != nil {
			return "KO", err
		}
		if res != nil {
			defer res.Body.Close()
		}
	}

	// Chesk response status.
	if res.StatusCode != http.StatusOK {
		return "KO", fmt.Errorf("http response status code: %v", res.Status)
	}

	// Chesk response body. The sentry health endpoint returns "ok" when there is no issue.
	var response []byte
	{
		var err error
		response, err = ioutil.ReadAll(res.Body)
		if err != nil {
			return "KO", err
		}
	}

	if strings.Compare(string(response), "ok") == 0 {
		return "OK", nil
	}

	return "KO", fmt.Errorf("response should be 'ok' but is: %v", string(response))
}
