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

func (m *mockFlakiModule) NextID(context.Context) (string, error) {
	m.nextIDCalled = true
	return "", nil
}

func (m *mockFlakiModule) NextValidID(context.Context) string {
	m.nextValidIDCalled = true
	return ""
}
