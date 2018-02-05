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
	var mockFlakiModule = &mockFlakiModule{}

	var srv = NewComponent(mockFlakiModule)

	// NextID
	var expectedID = strconv.FormatUint(rand.Uint64(), 10)
	mockFlakiModule.id = expectedID
	mockFlakiModule.nextIDCalled = false
	mockFlakiModule.fail = false
	var id, err = srv.NextID(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, expectedID, id)
	assert.True(t, mockFlakiModule.nextIDCalled)

	// NextID error.
	mockFlakiModule.id = strconv.FormatUint(rand.Uint64(), 10)
	mockFlakiModule.nextIDCalled = false
	mockFlakiModule.fail = true
	id, err = srv.NextID(context.Background())
	assert.NotNil(t, err)
	assert.Zero(t, id)
	assert.True(t, mockFlakiModule.nextIDCalled)

	// NextValidID.
	expectedID = strconv.FormatUint(rand.Uint64(), 10)
	mockFlakiModule.id = expectedID
	mockFlakiModule.nextValidIDCalled = false
	id = srv.NextValidID(context.Background())
	assert.Equal(t, expectedID, id)
	assert.True(t, mockFlakiModule.nextValidIDCalled)
}

// Mock Flaki module.
type mockFlakiModule struct {
	id                string
	nextIDCalled      bool
	nextValidIDCalled bool
	fail              bool
}

func (m *mockFlakiModule) NextID(context.Context) (string, error) {
	m.nextIDCalled = true
	if m.fail {
		return "", fmt.Errorf("fail")
	}
	return m.id, nil
}

func (m *mockFlakiModule) NextValidID(context.Context) string {
	m.nextValidIDCalled = true
	return m.id
}
