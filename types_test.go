package request_test

import (
	"fmt"
	"testing"

	"github.com/gildas/go-errors"
	"github.com/stretchr/testify/assert"
)

type failingReader int

func (r failingReader) Read(data []byte) (int, error) {
	return 0, errors.NotImplementedError.New()
}
func (r failingReader) Close() error {
	return nil
}

type stuff struct {
	ID string
}

func (s stuff) String() string {
	return s.ID
}

func TestStuffShouldBeStringer(t *testing.T) {
	s := stuff{"1234"}
	var z interface{} = s
	assert.NotNil(t, z.(fmt.Stringer), "Integer type is not a Stringer")

}