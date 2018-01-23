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
	var expected = rand.Uint64()
	var expectedStr = strconv.FormatUint(expected, 10)

	var flakiService = NewBasicService(&mockFlaki{
		id:   expected,
		fail: false,
	})

	var id, err = flakiService.NextID(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, expectedStr, id)
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
	var expectedStr = strconv.FormatUint(expected, 10)

	var flakiService = NewBasicService(&mockFlaki{
		id: expected,
	})

	var id = flakiService.NextValidID(context.Background())
	assert.Equal(t, expectedStr, id)
}

// Mock Flaki.
type mockFlaki struct {
	id   uint64
	fail bool
}

func (f *mockFlaki) NextID() (uint64, error) {
	if f.fail {
		return 0, fmt.Errorf("fail")
	}
	return f.id, nil
}

func (f *mockFlaki) NextIDString() (string, error) {
	var id, err = f.NextID()
	if err != nil {
		return "", err
	}
	return strconv.FormatUint(id, 10), nil
}

func (f *mockFlaki) NextValidID() uint64 {
	return f.id
}

func (f *mockFlaki) NextValidIDString() string {
	var id = f.NextValidID()
	return strconv.FormatUint(id, 10)
}
