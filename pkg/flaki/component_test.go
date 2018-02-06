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

func TestNewBasicService(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	var mockModule = &mockModule{}

	var srv = NewComponent(mockModule)

	// NextID
	var expectedID = strconv.FormatUint(rand.Uint64(), 10)
	mockModule.id = expectedID
	mockModule.nextIDCalled = false
	mockModule.fail = false
	var id, err = srv.NextID(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, expectedID, id)
	assert.True(t, mockModule.nextIDCalled)

	// NextID error.
	mockModule.id = strconv.FormatUint(rand.Uint64(), 10)
	mockModule.nextIDCalled = false
	mockModule.fail = true
	id, err = srv.NextID(context.Background())
	assert.NotNil(t, err)
	assert.Zero(t, id)
	assert.True(t, mockModule.nextIDCalled)

	// NextValidID.
	expectedID = strconv.FormatUint(rand.Uint64(), 10)
	mockModule.id = expectedID
	mockModule.nextValidIDCalled = false
	id = srv.NextValidID(context.Background())
	assert.Equal(t, expectedID, id)
	assert.True(t, mockModule.nextValidIDCalled)
}

// Mock Flaki module.
type mockModule struct {
	id                string
	nextIDCalled      bool
	nextValidIDCalled bool
	fail              bool
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
