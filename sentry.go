package main

import sentry "github.com/getsentry/raven-go"

// Sentry interface.
type Sentry interface {
	CaptureError(err error, tags map[string]string, interfaces ...sentry.Interface) string
	CaptureErrorAndWait(err error, tags map[string]string, interfaces ...sentry.Interface) string
	Close()
}

type NoopSentry struct{}

func (s *NoopSentry) CaptureError(err error, tags map[string]string, interfaces ...sentry.Interface) string {
	return ""
}
func (s *NoopSentry) CaptureErrorAndWait(err error, tags map[string]string, interfaces ...sentry.Interface) string {
	return ""
}
func (s *NoopSentry) Close() {}
