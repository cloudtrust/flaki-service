package main

import (
	"encoding/json"
)

type logstashLog struct {
	Timestamp       string            `json:"@timestamp"`
	LogstashVersion int               `json:"@version"`
	Fields          map[string]string `json:"@fields"`
	Message         string            `json:"@message, omitempty"`
}

type redisWriter struct {
	redis Redis
}

// Redis is the redis client interface.
type Redis interface {
	Send(commandName string, args ...interface{}) error
	Flush() error
	Close() error
}

// NewLogstashRedisWriter returns a writer that writes logs into a redis DB.
func NewLogstashRedisWriter(redis Redis) *redisWriter {
	return &redisWriter{
		redis: redis,
	}
}

// Write writes logs into a redis DB.
func (w *redisWriter) Write(p []byte) (int, error) {
	// The current logs are json formatted by the go-kit JSONLogger.
	var logs = decodeJSON(p)

	// Encode to logstash format.
	var logstashLog, err = logstashEncode(logs)
	if err != nil {
		return 0, err
	}

	err = w.redis.Send("RPUSH", "flaki-service", logstashLog)
	if err != nil {
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

	var timestamp = m["ts"]
	delete(m, "ts")
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
