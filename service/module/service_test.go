package flaki

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNextID(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	var expected = rand.Uint64()

	var flakiService = NewBasicService(&mockFlaki{
		id:   expected,
		fail: false,
	})

	var id, err = flakiService.NextID(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, expected, id)
}

func TestNextIDFail(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	var expected = rand.Uint64()

	var flakiService = NewBasicService(&mockFlaki{
		id:   expected,
		fail: true,
	})

	var id, err = flakiService.NextID(context.Background())
	assert.NotNil(t, err)
	assert.Zero(t, id)
}

func TestNextValidID(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	var expected = rand.Uint64()

	var flakiService = NewBasicService(&mockFlaki{
		id: expected,
	})

	var id = flakiService.NextValidID(context.Background())
	assert.Equal(t, expected, id)
}

type mockFlaki struct {
	id   uint64
	fail bool
}

func (m *mockFlaki) NextID() (uint64, error) {
	if m.fail {
		return 0, fmt.Errorf("fail")
	}
	return m.id, nil
}

func (m *mockFlaki) NextValidID() uint64 {
	return m.id
}
