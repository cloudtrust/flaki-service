package flaki

import (
	"context"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNextIDEndpoint(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var mockComponent = &mockComponent{fail: false, id: flakiID}

	// Context with correlation ID.
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", corrID)

	var e = MakeNextIDEndpoint(mockComponent)

	// NextID.
	var id, err = e(ctx, nil)
	assert.Nil(t, err)
	assert.Equal(t, flakiID, id)
}

func TestNextValidIDEndpoint(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var mockComponent = &mockComponent{id: flakiID}

	// Context with correlation ID.
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", corrID)

	var e = MakeNextValidIDEndpoint(mockComponent)

	// NextValidID.
	var id, err = e(ctx, nil)
	assert.Nil(t, err)
	assert.Equal(t, flakiID, id)
}
