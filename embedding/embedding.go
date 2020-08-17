package embedding

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/maja42/ember/internal"
)

const compatibleVersion = "maja42/ember/v1"

// PrintlnFunc is used for logging the embedding progress.
type PrintlnFunc func(format string, args ...interface{})

// Embed embeds the attachments into the target executable.
//
// out receives the target executable including all attachments.
//
// exe reads from the target executable that should be augmented.
// Embed verifies that the target executable is compatible with this version of ember
// by searching for the magic marker-string (compiled into every executable that imports ember).
// Embed fails if the executable is incompatible or already contains embedded content.
//
// attachments is a map of attachment names to the corresponding readers for the content.
//
// logger (optional) is used to report the progress during embedding.
//
// Note that all ReadSeekers are seeked to their start before usage,
// meaning the entirety of readable content is embedded. Use io.SectionReader to avoid this.
func Embed(out io.Writer, exe io.ReadSeeker, attachments map[string]io.ReadSeeker, logger PrintlnFunc) error {
	if logger == nil {
		logger = func(string, ...interface{}) {}
	}

	if err := verifyTargetExe(exe); err != nil {
		return fmt.Errorf("verify executable: %w", err)
	}

	toc, err := buildTOC(attachments)
	if err != nil {
		return fmt.Errorf("build TOC: %w", err)
	}
	jsonTOC, err := json.Marshal(toc)
	if err != nil {
		return fmt.Errorf("marshal TOC: %w", err)
	}

	// Executable
	logger("Writing executable")
	if _, err := io.Copy(out, exe); err != nil {
		return fmt.Errorf("copy executable: %w", err)
	}
	// Boundary
	if err := internal.WriteBoundary(out); err != nil {
		return err
	}
	// TOC
	logger("Adding TOC (%d bytes)", len(jsonTOC))
	if _, err := out.Write(jsonTOC); err != nil {
		return fmt.Errorf("write TOC: %w", err)
	}
	// Boundary
	if err := internal.WriteBoundary(out); err != nil {
		return err
	}
	// Attachments
	for _, att := range toc {
		logger("Adding %q (%d bytes)", att.Name, att.Size)
		if _, err := io.Copy(out, attachments[att.Name]); err != nil {
			return fmt.Errorf("write attachment %q: %w", att.Name, err)
		}
	}
	// Boundary
	if err := internal.WriteBoundary(out); err != nil {
		return err
	}
	return nil
}

// EmbedFiles embeds the given files into the target executable.
//
// attachments is a map of attachment names to the respective file's filepath.
//
// See Embed for more information.
func EmbedFiles(out io.Writer, exe io.ReadSeeker, attachments map[string]string, logger PrintlnFunc) error {
	reader := make(map[string]io.ReadSeeker, len(attachments))

	for name, path := range attachments {
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("open attachment %q (%q): %w", name, path, err)
		}
		//goland:noinspection ALL
		defer file.Close()
		reader[name] = file
	}
	return Embed(out, exe, reader, logger)
}

// verifyTargetExe ensures that the target executable is compatible.
// The reader is seeked to the beginning afterwards.
func verifyTargetExe(exe io.ReadSeeker) error {
	// Rewind seeker to start-of-executable (just in case)
	if _, err := exe.Seek(0, io.SeekStart); err != nil {
		return err
	}

	// Check if the target executable is compatible.
	// Compatible executables are importing 'ember' in the correct version,
	// causing a marker-string to be present in the binary.
	// String-replace is used to ensure the marker is not present in the embedder-executable.
	marker := "~~MagicMarker for XXX~~"
	marker = strings.ReplaceAll(marker, "XXX", compatibleVersion)

	offset := internal.SeekPattern(exe, []byte(marker))
	if offset == -1 { // not a go executable, or does not import correct library(-version)
		return errors.New("incompatible (magic string not found)")
	}

	offset = internal.SeekBoundary(exe)
	if offset != -1 {
		return errors.New("already contains embedded content")
	}

	if _, err := exe.Seek(0, io.SeekStart); err != nil {
		return err
	}
	return nil
}

// buildTOC returns the TOC (table-of-contents) for embedding the given data.
// All attachments are seeked to the beginning afterwards.
func buildTOC(attachments map[string]io.ReadSeeker) (internal.TOC, error) {
	toc := make(internal.TOC, 0, len(attachments))

	for name, r := range attachments {
		size, err := getSize(r)
		if err != nil {
			return nil, fmt.Errorf("attachment %q: %w", name, err)
		}
		toc = append(toc, internal.Attachment{
			Name: name,
			Size: size,
		})
	}
	return toc, nil
}

// getSize returns the size of the readable content.
// The reader is seeked to the beginning afterwards.
func getSize(r io.ReadSeeker) (int64, error) {
	size, err := r.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, err
	}
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return 0, err
	}
	return size, nil
}
