package flaki

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNextID(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	var expectedID = strconv.FormatUint(rand.Uint64(), 10)
	var mockFlaki = &mockFlaki{fail: false, id: expectedID}
	var m = NewModule(mockFlaki)

	var id, err = m.NextID(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, expectedID, id)
}

func TestNextIDFail(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	var expectedID = strconv.FormatUint(rand.Uint64(), 10)
	var mockFlaki = &mockFlaki{fail: true, id: expectedID}
	var m = NewModule(mockFlaki)

	// When an error is returned, the id is the zero string.
	var id, err = m.NextID(context.Background())
	assert.NotNil(t, err)
	assert.Zero(t, id)
}

func TestNextValidID(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	var expectedID = strconv.FormatUint(rand.Uint64(), 10)
	var mockFlaki = &mockFlaki{id: expectedID}
	var m = NewModule(mockFlaki)

	var id = m.NextValidID(context.Background())
	assert.Equal(t, expectedID, id)
}

// Mock Flaki.
type mockFlaki struct {
	id   string
	fail bool
}

func (f *mockFlaki) NextIDString() (string, error) {
	if f.fail {
		return "", fmt.Errorf("fail")
	}
	return f.id, nil
}

func (f *mockFlaki) NextValidIDString() string {
	return f.id
}

// Mock Flaki module.
type mockModule struct {
	id                string
	fail              bool
	nextIDCalled      bool
	nextValidIDCalled bool
}

func (m *mockModule) NextID(context.Context) (string, error) {
	m.nextIDCalled = true
	if m.fail {
		return "", fmt.Errorf("fail")
	}
	return m.id, nil
}

func (m *mockModule) NextValidID(context.Context) string {
	m.nextValidIDCalled = true
	return m.id
}
