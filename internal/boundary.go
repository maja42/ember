package internal

import (
	"bufio"
	"bytes"
	"io"
)

// boundaryPart is appended n-times after the unzipper executable and marks the boundary between the executable and attachments.
// It is used by the unzipper to detect when attachments start.
var boundaryPart = []byte{'#', 15, 1, 12, 1, '#'}

// boundaryPartCount defines how often the boundary-character is repeated.
// This ensures that the pattern does not appear by accident within the executable.
const boundaryPartCount = 4

// boundary is "boundaryPart" repeated "boundaryPartCount" times
var boundary []byte

// BoundarySize will contain the size of the complete boundary pattern
var BoundarySize int

func init() {
	// build up boundary in-memory
	partLen := len(boundaryPart)
	BoundarySize = partLen * boundaryPartCount

	boundary = make([]byte, BoundarySize)
	for i := 0; i < boundaryPartCount; i++ {
		copy(boundary[i*partLen:], boundaryPart)
	}
}

// IsBoundary checks if the given byte slice equals the boundary.
func IsBoundary(data []byte) bool {
	return bytes.Equal(boundary, data)
}

// WriteSection writes a new section with arbitrary data.
func WriteBoundary(w io.Writer) error {
	if _, err := w.Write(boundary); err != nil {
		return err
	}
	return nil
}

// SeekBoundary reads from the reader until the end of the boundary.
// Returns the number of bytes (offset) that were read (including the pattern itself).
// Returns -1 if the boundary was not found.
func SeekBoundary(in io.ReadSeeker) int64 {
	return SeekPattern(in, boundary)
}

// SeekPattern reads from the reader until the search pattern was found.
// The next byte coming from the reader will be the first byte after the pattern ended.
// Returns the number of bytes (offset) that were read (including the pattern itself).
// Returns -1 if the pattern was not found.
func SeekPattern(in io.ReadSeeker, pattern []byte) int64 {
	rPos, _ := in.Seek(0, io.SeekCurrent)

	var offset int64
	r := bufio.NewReader(in)

	nIdx := 0 // #bytes we already found
	for nIdx < len(pattern) {
		b, err := r.ReadByte()
		if err != nil { // not found
			return -1
		}
		if pattern[nIdx] == b {
			nIdx++
		} else {
			nIdx = 0
		}
		offset++
	}

	// seek the reader after the pattern (needed, because reading was done via the buffer)
	_, _ = in.Seek(rPos+offset, io.SeekStart)
	return offset
}
