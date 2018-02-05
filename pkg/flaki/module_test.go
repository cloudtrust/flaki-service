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

	var flakiService = NewModule(&mockFlaki{
		id:   expectedID,
		fail: false,
	})

	var id, err = flakiService.NextID(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, expectedID, id)
}

func TestNextIDFail(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	var flakiService = NewModule(&mockFlaki{
		id:   strconv.FormatUint(rand.Uint64(), 10),
		fail: true,
	})

	// When an error is returned, the id is the zero string.
	var id, err = flakiService.NextID(context.Background())
	assert.NotNil(t, err)
	assert.Zero(t, id)
}

func TestNextValidID(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	var expected = strconv.FormatUint(rand.Uint64(), 10)

	var flakiService = NewModule(&mockFlaki{
		id: expected,
	})

	var id = flakiService.NextValidID(context.Background())
	assert.Equal(t, expected, id)
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
