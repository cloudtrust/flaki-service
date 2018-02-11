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

	var expectedID = strconv.FormatUint(rand.Uint64(), 10)
	var mockModule = &mockModule{fail: false, id: expectedID}
	var c = NewComponent(mockModule)

	// NextID.
	mockModule.nextIDCalled = false
	var id, err = c.NextID(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, expectedID, id)
	assert.True(t, mockModule.nextIDCalled)

	// NextID error.
	mockModule.nextIDCalled = false
	mockModule.fail = true
	id, err = c.NextID(context.Background())
	assert.NotNil(t, err)
	assert.Zero(t, id)
	assert.True(t, mockModule.nextIDCalled)

	// NextValidID.
	expectedID = strconv.FormatUint(rand.Uint64(), 10)
	mockModule.id = expectedID
	mockModule.nextValidIDCalled = false
	id = c.NextValidID(context.Background())
	assert.Equal(t, expectedID, id)
	assert.True(t, mockModule.nextValidIDCalled)
}

// Mock component.
type mockComponent struct {
	id                string
	fail              bool
	nextIDCalled      bool
	nextValidIDCalled bool
}

func (c *mockComponent) NextID(context.Context) (string, error) {
	c.nextIDCalled = true
	if c.fail {
		return "", fmt.Errorf("fail")
	}
	return c.id, nil
}

func (c *mockComponent) NextValidID(context.Context) string {
	c.nextValidIDCalled = true
	return c.id
}
