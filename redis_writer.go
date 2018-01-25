package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/garyburd/redigo/redis"
	"github.com/go-kit/kit/log"
)

type logstashLog struct {
	Timestamp       string            `json:"@timestamp"`
	LogstashVersion int               `json:"@version"`
	Fields          map[string]string `json:"@fields"`
	Message         string            `json:"@message, omitempty"`
}

type redisWriter struct {
	pool *redis.Pool
}

func NewLogstashRedisWriter(pool *redis.Pool) io.Writer {
	var redisWriter = &redisWriter{
		pool: pool,
	}
	return redisWriter
}

func (w *redisWriter) Write(p []byte) (int, error) {
	// The current logs are json formatted by the go-kit JSONLogger.
	var logs = decodeJSON(p)

	// Encode to logstash format.
	var logstashLog, err = logstashEncode(logs)
	if err != nil {
		fmt.Println(err)
		return 0, err
	}

	// Get connection
	var c = w.pool.Get()
	defer c.Close()

	err = c.Send("RPUSH", "flaki-service", logstashLog)
	if err != nil {
		fmt.Println(err)
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

type MultiLogger interface {
	log.Logger
	AddLogger(logger log.Logger)
}

type multiLogger struct {
	loggers []log.Logger
}

func NewMultiLogger(loggers ...log.Logger) MultiLogger {
	var l = append([]log.Logger{}, loggers...)
	return &multiLogger{
		loggers: l,
	}
}

func (l *multiLogger) AddLogger(logger log.Logger) {
	l.loggers = append(l.loggers, logger)
}

func (l *multiLogger) Log(keyvals ...interface{}) error {
	for _, logger := range l.loggers {
		var err = logger.Log(keyvals...)
		if err != nil {
			return err
		}
	}
	return nil
}
