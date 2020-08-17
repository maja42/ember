package embedding

import (
	"io"
	"strings"
	"testing"

	"github.com/maja42/ember/internal"
	"github.com/stretchr/testify/assert"
)

func Test_buildTOC(t *testing.T) {
	r1 := strings.NewReader("content 1")
	r2 := strings.NewReader("second content")

	attachments := map[string]io.ReadSeeker{
		"first":  r1,
		"second": r2,
	}

	toc, err := buildTOC(attachments)
	assert.NoError(t, err)
	assert.Len(t, toc, 2)

	assert.Contains(t, toc, internal.Attachment{
		Name: "first",
		Size: 9,
	})
	assert.Contains(t, toc, internal.Attachment{
		Name: "second",
		Size: 14,
	})
}
