package flakie

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewEndpoints(t *testing.T) {
	// Context with correlation ID.
	rand.Seed(time.Now().UnixNano())
	var ctx = context.WithValue(context.Background(), "correlation_id", 0)

	var endpoints = NewEndpoints()

	// NextID.
	var expectedID = strconv.FormatUint(rand.Uint64(), 10)
	endpoints = endpoints.MakeNextIDEndpoint(&mockFlakiComponent{
		id:   expectedID,
		fail: false,
	},
	)
	var id, err = endpoints.NextID(ctx)
	assert.Nil(t, err)
	assert.Equal(t, expectedID, id)

	// NextValidID.
	expectedID = strconv.FormatUint(rand.Uint64(), 10)
	endpoints = endpoints.MakeNextValidIDEndpoint(&mockFlakiComponent{
		id:   expectedID,
		fail: false,
	},
	)
	id = endpoints.NextValidID(ctx)
	assert.Equal(t, expectedID, id)
}

// Mock Flaki Component.
type mockFlakiComponent struct {
	id   string
	fail bool
}

func (s *mockFlakiComponent) NextID(context.Context) (string, error) {
	if s.fail {
		return "", fmt.Errorf("fail")
	}
	return s.id, nil
}

func (s *mockFlakiComponent) NextValidID(context.Context) string {
	return s.id
}
