package endpoint

import (
	"context"
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
	endpoints = endpoints.MakeNextIDEndpoint(&mockFlakiService{
		id:   expectedID,
		fail: false,
	},
	)
	var id, err = endpoints.NextID(ctx)
	assert.Nil(t, err)
	assert.Equal(t, expectedID, id)

	// NextValidID.
	expectedID = strconv.FormatUint(rand.Uint64(), 10)
	endpoints = endpoints.MakeNextValidIDEndpoint(&mockFlakiService{
		id:   expectedID,
		fail: false,
	},
	)
	id = endpoints.NextValidID(ctx)
	assert.Equal(t, expectedID, id)
}
