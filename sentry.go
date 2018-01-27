package main

import sentry "github.com/getsentry/raven-go"

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

// Close does nothing for the receiver NoopSentry.
func (s *NoopSentry) Close() {}
