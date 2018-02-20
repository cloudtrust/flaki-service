package health

import (
	"context"
	"time"
)

// RedisModule is the health check module for redis.
type RedisModule interface {
	HealthChecks(context.Context) []redisHealthReport
}

type redisModule struct {
	redis Redis
}

type redisHealthReport struct {
	Name     string
	Duration string
	Status   Status
	Error    string
}

// Redis is the interface of the redis client.
type Redis interface {
	Do(cmd string, args ...interface{}) (interface{}, error)
}

// NewRedisModule returns the redis health module.
func NewRedisModule(redis Redis) RedisModule {
	return &redisModule{redis: redis}
}

// HealthChecks executes all health checks for Redis.
func (m *redisModule) HealthChecks(context.Context) []redisHealthReport {
	var reports = []redisHealthReport{}
	reports = append(reports, redisPingCheck(m.redis))
	return reports
}

func redisPingCheck(redis Redis) redisHealthReport {
	// If redis is deactivated.
	if redis == nil {
		return redisHealthReport{
			Name:     "ping",
			Duration: "N/A",
			Status:   Deactivated,
		}
	}

	var now = time.Now()
	var _, err = redis.Do("PING")
	var duration = time.Since(now)

	var status = OK
	var error = ""
	if err != nil {
		status = KO
		error = err.Error()
	}

	return redisHealthReport{
		Name:     "ping",
		Duration: duration.String(),
		Status:   status,
		Error:    error,
	}
}
