package main

import (
	"time"

	"github.com/go-kit/kit/metrics"
	gokit_influx "github.com/go-kit/kit/metrics/influx"
	influx "github.com/influxdata/influxdb/client/v2"
)

// No operation Influx client.
type NoopInflux struct{}

func (in *NoopInflux) Ping(timeout time.Duration) (time.Duration, string, error) { return 0, "", nil }
func (in *NoopInflux) Write(bp influx.BatchPoints) error                         { return nil }
func (in *NoopInflux) Query(q influx.Query) (*influx.Response, error) {
	return &influx.Response{
		Results: []influx.Result{},
		Err:     "",
	}, nil
}
func (in *NoopInflux) Close() error { return nil }

// No operation go-kit Influx.
type GokitInflux interface {
	NewCounter(name string) metrics.Counter
	NewGauge(name string) metrics.Gauge
	NewHistogram(name string) metrics.Histogram
	WriteLoop(c <-chan time.Time, w gokit_influx.BatchPointsWriter)
	WriteTo(w gokit_influx.BatchPointsWriter) (err error)
}

// NoopGokitInflux is a go-kit influx that does nothing.
type NoopGokitInflux struct{}

func (in *NoopGokitInflux) NewCounter(name string) metrics.Counter {
	return &NoopCounter{}
}
func (in *NoopGokitInflux) NewGauge(name string) metrics.Gauge {
	return &NoopGauge{}
}
func (in *NoopGokitInflux) NewHistogram(name string) metrics.Histogram {
	return &NoopHistogram{}
}
func (in *NoopGokitInflux) WriteLoop(c <-chan time.Time, w gokit_influx.BatchPointsWriter) {}
func (in *NoopGokitInflux) WriteTo(w gokit_influx.BatchPointsWriter) (err error)           { return nil }

// NoopGauge is a Counter that does nothing.
type NoopCounter struct{}

func (c *NoopCounter) With(labelValues ...string) metrics.Counter {
	return c
}
func (c *NoopCounter) Add(delta float64) {}

// NoopGauge is a Gauge that does nothing.
type NoopGauge struct{}

func (g *NoopGauge) With(labelValues ...string) metrics.Gauge {
	return g
}
func (g *NoopGauge) Set(value float64) {}
func (g *NoopGauge) Add(delta float64) {}

// NoopGauge is an Histogram that does nothing.
type NoopHistogram struct{}

func (h *NoopHistogram) With(labelValues ...string) metrics.Histogram {
	return h
}
func (h *NoopHistogram) Observe(value float64) {}
