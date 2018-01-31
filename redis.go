package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/garyburd/redigo/redis"
)

type logstashLog struct {
	Timestamp       string            `json:"@timestamp"`
	LogstashVersion int               `json:"@version"`
	Fields          map[string]string `json:"@fields"`
	Message         string            `json:"@message, omitempty"`
}

type redisWriter struct {
	con redis.Conn
}

func NewLogstashRedisWriter(con redis.Conn) *redisWriter {
	return &redisWriter{
		con: con,
	}
}

func (w *redisWriter) Write(p []byte) (int, error) {
	// The current logs are json formatted by the go-kit JSONLogger.
	var logs = decodeJSON(p)

	// Encode to logstash format.
	var logstashLog, err = logstashEncode(logs)
	if err != nil {
		return 0, err
	}

	err = w.con.Send("RPUSH", "flaki-service", logstashLog)
	if err != nil {
		fmt.Printf("redis err: %s", err)
		return 0, err
	}
	return len(p), nil
}

func decodeJSON(d []byte) map[string]string {
	var logs = make(map[string]string)
	json.Unmarshal(d, &logs)
	return logs
}

func logstashEncode(m map[string]string) ([]byte, error) {

	var timestamp = m["time"]
	delete(m, "time")
	var msg = m["msg"]
	delete(m, "msg")

	var l = logstashLog{
		Timestamp:       timestamp,
		LogstashVersion: 1,
		Fields:          m,
		Message:         msg,
	}

	var err error
	var ll []byte
	ll, err = json.Marshal(l)
	return ll, err
}

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
