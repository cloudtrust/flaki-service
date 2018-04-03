package health

//go:generate mockgen -destination=./mock/redis.go -package=mock -mock_names=RedisClient=RedisClient  github.com/cloudtrust/flaki-service/pkg/health RedisClient

import (
	"context"
	"time"

	"github.com/pkg/errors"
)

// RedisModule is the health check module for redis.
type RedisModule struct {
	redis   RedisClient
	enabled bool
}

// RedisClient is the interface of the redis client.
type RedisClient interface {
	Do(cmd string, args ...interface{}) (interface{}, error)
}

// NewRedisModule returns the redis health module.
func NewRedisModule(redis RedisClient, enabled bool) *RedisModule {
	return &RedisModule{
		redis:   redis,
		enabled: enabled,
	}
}

// RedisReport is the health report returned by the redis module.
type RedisReport struct {
	Name     string
	Duration time.Duration
	Status   Status
	Error    error
}

// HealthChecks executes all health checks for Redis.
func (m *RedisModule) HealthChecks(context.Context) []RedisReport {
	var reports = []RedisReport{}
	reports = append(reports, m.redisPingCheck())
	return reports
}

func (m *RedisModule) redisPingCheck() RedisReport {
	var healthCheckName = "ping"

	if !m.enabled {
		return RedisReport{
			Name:   healthCheckName,
			Status: Deactivated,
		}
	}

	var now = time.Now()
	var _, err = m.redis.Do("PING")
	var duration = time.Since(now)

	var hcErr error
	var s Status
	switch {
	case err != nil:
		hcErr = errors.Wrap(err, "could not ping redis")
		s = KO
	default:
		s = OK
	}

	return RedisReport{
		Name:     healthCheckName,
		Duration: duration,
		Status:   s,
		Error:    hcErr,
	}
}
