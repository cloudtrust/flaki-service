package health

import (
	"context"
	"time"
)

type RedisHealthReport struct {
	Name     string
	Duration string
	Status   string
	Error    string
}

type Redis interface {
	Do(cmd string, args ...interface{}) (interface{}, error)
}

type RedisHealthModule struct {
	redis Redis
}

// NewRedisHealthModule returns the redis health module.
func NewRedisHealthModule(redis Redis) *RedisHealthModule {
	return &RedisHealthModule{redis: redis}
}

// HealthChecks executes all health checks for Redis.
func (s *RedisHealthModule) HealthChecks(context.Context) []RedisHealthReport {
	var reports = []RedisHealthReport{}
	reports = append(reports, redisPingCheck(s.redis))
	return reports
}

func redisPingCheck(redis Redis) RedisHealthReport {
	var now = time.Now()
	var _, err = redis.Do("PING")
	var duration = time.Since(now)

	var status = "OK"
	var error = ""
	if err != nil {
		status = "KO"
		error = err.Error()
	}

	return RedisHealthReport{
		Name:     "ping",
		Duration: duration.String(),
		Status:   status,
		Error:    error,
	}
}
