package health

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

type Health struct {
	healthChecks []*HealthChecks
}

func (h *Health) AddCheck(check *HealthChecks) {
	h.healthChecks = append(h.healthChecks, check)
}

func (h *Health) RegisterRoutes(route *mux.Router) {
	var healthSubroute = route.PathPrefix("/health").Subrouter()
	healthSubroute.Handle("", h.HealthChecksRoute())

	// Subroutes.
	for _, check := range h.healthChecks {
		var routeName = fmt.Sprintf("/%s", check.name)

		healthSubroute.HandleFunc(routeName, http.HandlerFunc(MakeHandler(check)))
	}
}

func (h *Health) HealthChecksRoute() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var report = map[string]string{}

		for _, check := range h.healthChecks {
			report[check.name] = check.Status()
		}

		var j, err = json.MarshalIndent(report, "", "  ")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write(j)
		}
	}
}
func MakeHandler(check *HealthChecks) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var reports = check.Report()

		var d, err = json.MarshalIndent(reports, "", "  ")

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write(d)
		}
	}
}

type HealthChecks struct {
	name   string
	checks []HealthTest
}

func (hc *HealthChecks) Status() string {
	var status = "OK"
	for _, tst := range hc.checks {
		var report = tst()
		if report.Status != "OK" {
			status = "KO"
		}
	}
	return status
}

func (hc *HealthChecks) Report() []TestReport {

	var reports = []TestReport{}

	for _, tst := range hc.checks {
		reports = append(reports, tst())
	}

	return reports
}

type TestReport struct {
	Name     string `json:"name"`
	Duration string `json:"duration"`
	Status   string `json:"status"`
}

type HealthTest func() TestReport

// INFLUX
/*
func MakeInfluxHealthChecks(client influx.Client) *HealthChecks {

	var checks = []HealthTest{
		makePingHealthTest(client),
	}

	return &HealthChecks{
		name:   "influx",
		checks: checks,
	}
}

func makePingHealthTest(client influx.Client) HealthTest {
	return func() TestReport {
		var d, _, err = client.Ping(time.Duration(5 * time.Second))
		var status = "OK"
		if err != nil {
			status = "KO"
		}

		return TestReport{
			Name:     "ping",
			Duration: d.String(),
			Status:   status,
		}
	}
}


// Redis

func MakeRedisHealthChecks(pool *redis.Pool) *HealthChecks {

	var checks = []HealthTest{
		makeRedisPingHealthTest(pool),
	}

	return &HealthChecks{
		name:   "redis",
		checks: checks,
	}
}

func makeRedisPingHealthTest(pool *redis.Pool) HealthTest {
	return func() TestReport {
		// Get connection
		var c = pool.Get()
		defer c.Close()

		var now = time.Now()
		var _, err = c.Do("PING")
		var duration = time.Since(now)

		var status = "OK"
		if err != nil {
			status = "KO"
		}

		return TestReport{
			Name:     "ping",
			Duration: duration.String(),
			Status:   status,
		}
	}
}

// Sentry

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
*/
