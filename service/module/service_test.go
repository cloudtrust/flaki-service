package flaki

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

type mockFlaki struct {
	id   uint64
	fail bool
}

var ErrFail = errors.New("Fail")

func (m *mockFlaki) NextID() (uint64, error) {
	if m.fail {
		return 0, ErrFail
	}
	return m.id, nil
}

func (m *mockFlaki) NextValidID() uint64 {
	return m.id
}

func TestBasicService_NextID(t *testing.T) {
	var expected uint64 = 1
	var flakiService = NewBasicService(&mockFlaki{
		id:   expected,
		fail: false,
	})

	var id, err = flakiService.NextID(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, expected, id)
}

func TestBasicService_NextIDFail(t *testing.T) {
	var expected uint64 = 2

	var flakiService = NewBasicService(&mockFlaki{
		id:   expected,
		fail: true,
	})

	var id, err = flakiService.NextID(context.Background())
	assert.NotNil(t, err)
	assert.Zero(t, id)
}

func TestBasicService_NextValidID(t *testing.T) {
	var expected uint64 = 3

	var flakiService = NewBasicService(&mockFlaki{
		id: expected,
	})

	var id = flakiService.NextValidID(context.Background())
	assert.Equal(t, expected, id)
}
