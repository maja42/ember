package internal

import (
	"bytes"
	"crypto/rand"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitBoundary(t *testing.T) {
	assert.Equal(t, BoundarySize, len(boundaryPart)*boundaryPartCount)

	assert.Len(t, boundary, len(boundaryPart)*boundaryPartCount)
	for i := 0; i < boundaryPartCount; i++ {
		b := boundary[i*len(boundaryPart):]
		assert.Equal(t, boundaryPart, b[:len(boundaryPart)])
	}
}

func TestIsBoundary(t *testing.T) {
	assert.True(t, IsBoundary(boundary[:]))

	assert.False(t, IsBoundary(boundary[1:]))

	moreBoundary := append(boundary, 0)
	assert.False(t, IsBoundary(moreBoundary))

}

func TestWriteBoundary(t *testing.T) {
	buf := new(bytes.Buffer)
	err := WriteBoundary(buf)
	assert.NoError(t, err)

	assert.Equal(t, len(boundaryPart)*boundaryPartCount, buf.Len())

	for i := 0; i < boundaryPartCount; i++ {
		b := buf.Bytes()[i*len(boundaryPart):]
		assert.Equal(t, boundaryPart, b[:len(boundaryPart)])
	}
}

type errWriter struct{}

func (errWriter) Write([]byte) (n int, err error) {
	return 0, errors.New("simulated error")
}

func TestWriteBoundary_writeError(t *testing.T) {
	err := WriteBoundary(errWriter{})
	assert.EqualError(t, err, "simulated error")
}

func TestSeekBoundary(t *testing.T) {
	// Create buffer:
	//	- random bytes
	//	- boundary
	//  - "text 1"
	//	- boundary
	//  - "text 2"
	randomBytes := 50
	random := make([]byte, randomBytes)
	_, err := rand.Read(random)
	assert.NoError(t, err)

	buf := bytes.NewBuffer(random)
	buf.Write(boundary)
	buf.WriteString("text 1")
	buf.Write(boundary)
	buf.WriteString("text 2")

	r := bytes.NewReader(buf.Bytes())

	// seek first occurrence:
	offset := SeekBoundary(r)
	assert.Equal(t, int64(randomBytes+len(boundary)), offset)

	txt := make([]byte, 6)
	_, err = r.Read(txt)
	assert.NoError(t, err)
	assert.Equal(t, txt, []byte("text 1"))

	// seek second occurrence:
	offset = SeekBoundary(r)
	assert.Equal(t, int64(len(boundary)), offset)

	content, err := io.ReadAll(r)
	assert.NoError(t, err)
	assert.Equal(t, content, []byte("text 2"))

	// seek further:
	offset = SeekBoundary(r)
	assert.Equal(t, int64(-1), offset)
}

func TestSeekBoundary_noBoundary(t *testing.T) {
	randomBytes := 50
	random := make([]byte, randomBytes)
	_, err := rand.Read(random)
	assert.NoError(t, err)

	r := bytes.NewReader(random)

	// seek first occurrence:
	offset := SeekBoundary(r)
	assert.Equal(t, int64(-1), offset)
}
