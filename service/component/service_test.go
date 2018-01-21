package component

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBasicService(t *testing.T) {
	var mockFlakiModule = &mockFlakiModule{
		nextIDCalled:      false,
		nextValidIDCalled: false,
	}

	var srv = NewBasicService(mockFlakiModule)

	// NextID
	assert.False(t, mockFlakiModule.nextIDCalled)
	srv.NextID(context.Background())
	assert.True(t, mockFlakiModule.nextIDCalled)

	// NextValidID
	assert.False(t, mockFlakiModule.nextValidIDCalled)
	srv.NextValidID(context.Background())
	assert.True(t, mockFlakiModule.nextValidIDCalled)
}

type mockFlakiModule struct {
	nextIDCalled      bool
	nextValidIDCalled bool
}

func (m *mockFlakiModule) NextID(context.Context) (uint64, error) {
	m.nextIDCalled = true
	return 0, nil
}

func (m *mockFlakiModule) NextValidID(context.Context) uint64 {
	m.nextValidIDCalled = true
	return 0
}
