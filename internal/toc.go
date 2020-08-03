package internal

// TOC (=table of content) lists all attachments of an executable.
// The order of attachments in the TOC reflects the order of attachment data afterwards.
// The TOC is embedded as json prior to the first attachment, guarded by a boundary byte-pattern on both sides.
type TOC []Attachment

// Attachment represents a single embedded resource.
type Attachment struct {
	Name string // Resource name
	Size int64  // Resource size in bytes
}
