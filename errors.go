package ember

import "fmt"

// AttErr reports problems with embedded attachments.
type AttErr string

func (o *AttErr) Error() string {
	return string(*o)
}

func newAttErr(format string, a ...interface{}) *AttErr {
	err := AttErr(fmt.Sprintf(format, a...))
	return &err
}
