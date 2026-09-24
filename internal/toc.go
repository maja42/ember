package internal

// TOC (=table of content) lists all attachments of an executable.
// The TOC is embedded as json prior to the first attachment, guarded by a boundary byte-pattern on both sides.
//
// Entries are ordered by name, and the attachment data that follows is in that
// same order. This keeps the embedded output the same on every run, so
// executables built from the same attachments are reproducible.
//
// Readers must not rely on entries being sorted. They must look them up by name instead.
type TOC []Attachment

// Attachment represents a single embedded resource.
type Attachment struct {
	Name string // Resource name
	Size int64  // Resource size in bytes
}
