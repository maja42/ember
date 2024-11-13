package ember

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/maja42/ember/internal"
)

// Attachments represent embedded data in an executable.
type Attachments struct {
	exeFile *os.File
	offsets map[string]int64
	sizes   map[string]int64
}

// Open returns the attachments of the running executable.
func Open() (*Attachments, error) {
	path, err := os.Executable()
	if err != nil {
		return nil, err
	}
	if p, err := filepath.EvalSymlinks(path); err == nil {
		// EvalSymlinks fails on Windows if the executable is located in the
		// remote SYSVOL volume from the domain controller.
		// It is therefore optional, any errors are ignored.
		path = p
	}
	return OpenExe(path)
}

// OpenExe returns the attachments of an arbitrary executable.
func OpenExe(exePath string) (*Attachments, error) {
	att := &Attachments{}

	exe, err := os.Open(exePath)
	if err != nil {
		return nil, err
	}
	att.exeFile = exe
	dontClose := false
	defer func() {
		if !dontClose {
			_ = exe.Close()
		}
	}()

	// determine TOC location
	tocOffset := internal.SeekBoundary(exe)
	if tocOffset < 0 { // No attachments found
		dontClose = true
		return att, nil
	}
	nextBoundary := internal.SeekBoundary(exe)
	if nextBoundary < 0 {
		// first boundary was found, but the next one (indicating the end of TOC data) is missing.
		return nil, newAttErr("corrupt attachment data (incomplete TOC)")
	}
	tocEndOffset := tocOffset + nextBoundary
	tocSize := int(nextBoundary) - internal.BoundarySize

	// read TOC
	if _, err := exe.Seek(tocOffset, io.SeekStart); err != nil {
		return nil, err
	}

	var jsonTOC = make([]byte, tocSize)
	if _, err := io.ReadFull(exe, jsonTOC); err != nil {
		return nil, err
	}

	var toc internal.TOC
	if err := json.Unmarshal(jsonTOC, &toc); err != nil {
		return nil, newAttErr("corrupt attachment data (invalid TOC)")
	}

	// calc offsets
	att.offsets = make(map[string]int64, len(toc))
	att.sizes = make(map[string]int64, len(toc))
	offset := tocEndOffset
	for _, a := range toc {
		att.offsets[a.Name] = offset
		att.sizes[a.Name] = a.Size
		offset += a.Size
	}

	// find trailing boundary
	if _, err := exe.Seek(offset, io.SeekStart); err != nil {
		return nil, err
	}

	var trailer = make([]byte, internal.BoundarySize)
	if _, err := io.ReadFull(exe, trailer); err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF { // offsets point outside executable (missing data?)
			return nil, newAttErr("corrupt attachment data (offsets too large)")
		}
		return nil, err
	}
	if !internal.IsBoundary(trailer) {
		return nil, newAttErr("corrupt attachment data (invalid offsets)")
	}

	dontClose = true
	return att, nil
}

// Close the executable containing the attachments.
// Close will return an error if it has already been called.
func (a *Attachments) Close() error {
	return a.exeFile.Close()
}

// List returns a list containing the names of all attachments.
func (a *Attachments) List() []string {
	if len(a.offsets) == 0 { // no attachments
		return nil
	}
	l := make([]string, len(a.offsets))
	i := 0
	for name := range a.offsets {
		l[i] = name
		i++
	}
	return l
}

// Count returns the number of attachments.
func (a *Attachments) Count() int {
	return len(a.offsets)
}

// Reader groups basic methods available on attachments.
type Reader interface {
	io.ReadSeeker
	io.ReaderAt
	Size() int64
}

// Reader returns a reader for a given attachment.
// Returns nil if no attachment with that name exists.
func (a *Attachments) Reader(name string) Reader {
	offset, ok := a.offsets[name]
	if !ok {
		return nil
	}
	return io.NewSectionReader(a.exeFile, offset, a.sizes[name])
}

// Size returns the size of a specific attachment in bytes.
// Returns zero if no attachment with that name exists.
func (a *Attachments) Size(name string) int64 {
	return a.sizes[name]
}

// Offset returns the offset of a specific attachment in bytes, in relation to the start of the go executable.
// Returns zero if no attachment with that name exists.
func (a *Attachments) Offset(name string) int64 {
	return a.offsets[name]
}
