package main

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
