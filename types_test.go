package request_test

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/gildas/go-errors"
	"github.com/stretchr/testify/assert"
)

type failingReader int

func (r failingReader) Read(data []byte) (int, error) {
	return 0, errors.NotImplemented.WithStack()
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

type progressWriter struct {
	Total int64
	Max   int64
}

func (w *progressWriter) Write(p []byte) (n int, err error) {
	w.Total += int64(len(p))
	return len(p), nil
}

func (w *progressWriter) Close() error {
	return nil
}

type progressWriter2 struct {
	Total int64
	Max   int64
}

func (w *progressWriter2) SetMax64(max int64) {
	w.Max = max
}

func (w *progressWriter2) Write(p []byte) (n int, err error) {
	w.Total += int64(len(p))
	return len(p), nil
}

type progressWriter3 struct {
	Total int64
	Max   int64
}

func (w *progressWriter3) ChangeMax64(max int64) {
	w.Max = max
}

func (w *progressWriter3) Write(p []byte) (n int, err error) {
	w.Total += int64(len(p))
	return len(p), nil
}

func TestStuffShouldBeStringer(t *testing.T) {
	s := stuff{"1234"}
	var z interface{} = s
	assert.NotNil(t, z.(fmt.Stringer), "Integer type is not a Stringer")
}

// smallPNG returns a small PNG image as a byte array
//
// This is a 1x1 pixel PNG image and is 408 bytes in size
func smallPNG() []byte {
	image := "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAACklEQVR4nGMAAQAABQABDQottAAAAABJRU5ErkJggg=="
	data, err := base64.StdEncoding.DecodeString(image)
	if err != nil {
		panic(err)
	}
	return data
}
