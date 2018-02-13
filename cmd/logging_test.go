package main

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	jsonLog = "{\"msg\":\"logstash log\",\"caller\":\"flakid.go:120\",\"component_name\":\"flaki-service\",\"component_version\":\"1.0.0\",\"environment\":\"DEV\",\"git_commit\":\"5fb7de0d7ae3f3d5f5d6a322b2344bdab645fd33\",\"ts\":\"2018-02-13T06:27:07.123915229Z\"}"
)

func TestLogstashRedisWriter(t *testing.T) {
	var mockRedis = &mockRedis{}
	var w = NewLogstashRedisWriter(mockRedis)

	w.Write([]byte(jsonLog))

	var m = map[string]interface{}{}
	json.Unmarshal(mockRedis.logs, &m)

	assert.Equal(t, "2018-02-13T06:27:07.123915229Z", m["@timestamp"])
	assert.Equal(t, float64(1), m["@version"])
	assert.Equal(t, "logstash log", m["@message"])

	var fields = m["@fields"].(map[string]interface{})
	assert.Equal(t, "flakid.go:120", fields["caller"])
	assert.Equal(t, "flaki-service", fields["component_name"])
	assert.Equal(t, "1.0.0", fields["component_version"])
	assert.Equal(t, "DEV", fields["environment"])
	assert.Equal(t, "5fb7de0d7ae3f3d5f5d6a322b2344bdab645fd33", fields["git_commit"])
	var _, ok = fields["ts"]
	assert.False(t, ok)
	_, ok = fields["msg"]
	assert.False(t, ok)
}

func TestDecodeJSON(t *testing.T) {
	var m = decodeJSON([]byte(jsonLog))
	assert.Equal(t, "logstash log", m["msg"])
	assert.Equal(t, "flakid.go:120", m["caller"])
	assert.Equal(t, "flaki-service", m["component_name"])
	assert.Equal(t, "1.0.0", m["component_version"])
	assert.Equal(t, "DEV", m["environment"])
	assert.Equal(t, "5fb7de0d7ae3f3d5f5d6a322b2344bdab645fd33", m["git_commit"])
	assert.Equal(t, "2018-02-13T06:27:07.123915229Z", m["ts"])
}

func TestLogstashEncode(t *testing.T) {
	var logstashLog, err = logstashEncode(decodeJSON([]byte(jsonLog)))
	assert.Nil(t, err)

	var m = map[string]interface{}{}
	json.Unmarshal(logstashLog, &m)

	assert.Equal(t, "2018-02-13T06:27:07.123915229Z", m["@timestamp"])
	assert.Equal(t, float64(1), m["@version"])
	assert.Equal(t, "logstash log", m["@message"])

	var fields = m["@fields"].(map[string]interface{})
	assert.Equal(t, "flakid.go:120", fields["caller"])
	assert.Equal(t, "flaki-service", fields["component_name"])
	assert.Equal(t, "1.0.0", fields["component_version"])
	assert.Equal(t, "DEV", fields["environment"])
	assert.Equal(t, "5fb7de0d7ae3f3d5f5d6a322b2344bdab645fd33", fields["git_commit"])
	var _, ok = fields["ts"]
	assert.False(t, ok)
	_, ok = fields["msg"]
	assert.False(t, ok)
}

// Mock Redis.
type mockRedis struct {
	logs []byte
}

func (r *mockRedis) Send(commandName string, args ...interface{}) error {
	r.logs = args[1].([]byte)
	return nil
}

func (r *mockRedis) Flush() error { return nil }
func (r *mockRedis) Close() error { return nil }
