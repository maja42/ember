package embedding

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/maja42/ember"
	"github.com/maja42/ember/internal"
	"github.com/stretchr/testify/assert"
)

func TestEmbed(t *testing.T) {
	var out bytes.Buffer
	exe := prepareExecutableData()
	attachments := map[string]io.ReadSeeker{
		"att1": strings.NewReader("first content"),
		"att2": strings.NewReader("second content"),
	}

	err := Embed(&out, strings.NewReader(exe), attachments, nil)
	assert.NoError(t, err)

	out.WriteString("some other content attached later via another 3rd-party application")

	// verify executable content
	exeBytes := out.Bytes()[:len(exe)]
	assert.Equal(t, []byte(exe), exeBytes)

	// write content to disk
	tmpFile, err := os.CreateTemp("", "")
	assert.NoError(t, err)
	io.Copy(tmpFile, &out)
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	// verify 'ember' can extract the embedded attachments
	att, err := ember.OpenExe(tmpFile.Name())
	assert.NoError(t, err)
	assert.Equal(t, 2, att.Count())
	assert.Equal(t, int64(13), att.Size("att1"))
	assert.Equal(t, int64(14), att.Size("att2"))

	content, err := io.ReadAll(att.Reader("att1"))
	assert.NoError(t, err)
	assert.Equal(t, []byte("first content"), content)

	content, err = io.ReadAll(att.Reader("att2"))
	assert.NoError(t, err)
	assert.Equal(t, []byte("second content"), content)

	err = att.Close()
	assert.NoError(t, err)
}

func Test_verifyTargetExe(t *testing.T) {
	r := strings.NewReader(prepareExecutableData())
	err := verifyTargetExe(r, false)
	assert.Nil(t, err)
}

func Test_verifyTargetExe_invalidFile(t *testing.T) {
	r := strings.NewReader("does not contain magic marker")
	err := verifyTargetExe(r, false)
	assert.EqualError(t, err, "incompatible (magic string not found)")
}

func Test_verifyTargetExe_invalidFile_checkSkipped(t *testing.T) {
	r := strings.NewReader("does not contain magic marker")
	err := verifyTargetExe(r, true)
	assert.NoError(t, err)
}

func Test_verifyTargetExe_alreadyAugmented(t *testing.T) {
	var buf bytes.Buffer
	_ = internal.WriteBoundary(&buf)

	content := prepareExecutableData()
	content += buf.String()
	content += "some more content"

	r := strings.NewReader(content)

	err := verifyTargetExe(r, false)
	assert.EqualError(t, err, "already contains embedded content")
}

// The embedded output must not depend on the order of the map it was built
// from, which is what makes it reproducible. Both the TOC and the attachment
// data that follows it are ordered by name.
func TestEmbed_writesAttachmentsOrderedByName(t *testing.T) {
	// Each attachment has content of its own length, so any other order
	// produces different bytes.
	exe := prepareExecutableData()
	attachments := map[string]io.ReadSeeker{
		"delta":   strings.NewReader("dddd"),
		"bravo":   strings.NewReader("bb"),
		"alpha":   strings.NewReader("a"),
		"charlie": strings.NewReader("ccc"),
	}

	var out bytes.Buffer
	err := Embed(&out, strings.NewReader(exe), attachments, nil)
	assert.NoError(t, err)

	embedded := out.Bytes()[len(exe):]

	// The TOC sits between the first and second boundary.
	tocOffset := internal.SeekBoundary(bytes.NewReader(embedded))
	tocEnd := tocOffset + internal.SeekBoundary(bytes.NewReader(embedded[tocOffset:]))

	var toc internal.TOC
	err = json.Unmarshal(embedded[tocOffset:tocEnd-int64(internal.BoundarySize)], &toc)
	assert.NoError(t, err)

	names := make([]string, 0, len(toc))
	for _, att := range toc {
		names = append(names, att.Name)
	}
	assert.Equal(t, []string{"alpha", "bravo", "charlie", "delta"}, names)

	// The attachment data follows, in the same order.
	const expectedData = "abbcccdddd"
	data := embedded[tocEnd : tocEnd+int64(len(expectedData))]
	assert.Equal(t, expectedData, string(data))
}

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

func prepareExecutableData() string {
	randBytes := make([]byte, 0, 100)
	if _, err := rand.Read(randBytes); err != nil {
		panic(err)
	}

	exeData := "This is the executable content"
	exeData += string(randBytes)
	exeData += "~~MagicMarker for XXX~~"
	exeData += "Some more content"
	exeData += string(randBytes)

	exeData = strings.ReplaceAll(exeData, "XXX", compatibleVersion)
	return exeData
}
