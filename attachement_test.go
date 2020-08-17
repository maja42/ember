package ember

import (
	"crypto/rand"
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/maja42/ember/internal"
	"github.com/stretchr/testify/assert"
)

func TestAttachments(t *testing.T) {
	var testAttachments = [][]byte{
		[]byte("att1"),
		[]byte("2"),
		{},
		{0, 1, 2, 3},
	}

	var testTOC = internal.TOC{
		internal.Attachment{
			Name: "att1",
			Size: int64(len(testAttachments[0])),
		},
		internal.Attachment{
			Name: "num2",
			Size: int64(len(testAttachments[1])),
		},
		internal.Attachment{
			Name: "3",
			Size: int64(len(testAttachments[2])),
		},
		internal.Attachment{
			Name: "four",
			Size: int64(len(testAttachments[3])),
		},
	}

	path := prepareFile(t, testTOC, testAttachments)
	defer os.Remove(path)

	att, err := OpenExe(path)
	assert.NoError(t, err)

	assert.Len(t, att.offsets, len(testTOC))
	assert.Len(t, att.sizes, len(testTOC))

	t.Run("List()", func(t *testing.T) {
		list := att.List()
		assert.Len(t, list, len(testTOC))

		for _, name := range []string{"att1", "num2", "3", "four"} {
			assert.Contains(t, list, name)
		}
	})

	t.Run("Count()", func(t *testing.T) {
		count := att.Count()
		assert.Equal(t, len(testTOC), count)
	})

	t.Run("Reader(): success", func(t *testing.T) {
		r := att.Reader("att1")
		data, err := ioutil.ReadAll(r)
		assert.NoError(t, err)
		assert.Equal(t, string(testAttachments[0]), string(data))

		r = att.Reader("num2")
		data, err = ioutil.ReadAll(r)
		assert.NoError(t, err)
		assert.Equal(t, string(testAttachments[1]), string(data))

		r = att.Reader("3")
		data, err = ioutil.ReadAll(r)
		assert.NoError(t, err)
		assert.Equal(t, string(testAttachments[2]), string(data))

		r = att.Reader("four")
		data, err = ioutil.ReadAll(r)
		assert.NoError(t, err)
		assert.Equal(t, string(testAttachments[3]), string(data))
	})

	t.Run("Reader(): non-existing file", func(t *testing.T) {
		r := att.Reader("unknown")
		assert.Nil(t, r)
	})

	t.Run("Close()", func(t *testing.T) {
		err = att.Close()
		assert.NoError(t, err)
	})
}

func prepareFile(t *testing.T, toc internal.TOC, attachments [][]byte) string {
	file, err := ioutil.TempFile("", "")
	assert.NoError(t, err)
	defer file.Close()

	// write random data (=represents executable)
	random := make([]byte, 100)
	_, err = rand.Read(random)
	assert.NoError(t, err)
	_, err = file.Write(random)
	assert.NoError(t, err)

	// write boundary
	err = internal.WriteBoundary(file)
	assert.NoError(t, err)

	// write toc
	jsonTOC, err := json.Marshal(toc)
	assert.NoError(t, err)

	_, err = file.Write(jsonTOC)
	assert.NoError(t, err)

	// write boundary
	err = internal.WriteBoundary(file)
	assert.NoError(t, err)

	// write attachments
	for _, attachment := range attachments {
		_, err = file.Write(attachment)
		assert.NoError(t, err)
	}

	// write boundary
	err = internal.WriteBoundary(file)
	assert.NoError(t, err)

	// write random data (=represents trailing data)
	_, err = rand.Read(random)
	assert.NoError(t, err)
	_, err = file.Write(random)
	assert.NoError(t, err)

	filename := file.Name()
	return filename
}

func TestAttachments_NoAttachments(t *testing.T) {
	// Open the test executable, which should definitely not contain any attachments.
	att, err := Open()
	assert.NoError(t, err)

	list := att.List()
	assert.Nil(t, list)

	count := att.Count()
	assert.Zero(t, count)

	assert.Nil(t, att.Close())
}

func TestOpenExe_NoSuchFile(t *testing.T) {
	att, err := OpenExe("./:this file does not exist!")
	assert.Error(t, err)
	_, ok := err.(*os.PathError)
	assert.True(t, ok)
	assert.Nil(t, att)
}

func TestOpenExe_SecondBoundaryMissing(t *testing.T) {
	file, err := ioutil.TempFile("", "")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	random := make([]byte, 100)
	rand.Read(random)
	file.Write(random)
	internal.WriteBoundary(file)
	file.Close()

	att, err := OpenExe(file.Name())
	assert.EqualError(t, err, "corrupt attachment data (incomplete TOC)")
	assert.Nil(t, att)
}

func TestOpenExe_BrokenTOC(t *testing.T) {
	file, err := ioutil.TempFile("", "")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	random := make([]byte, 100)
	rand.Read(random)
	file.Write(random)
	internal.WriteBoundary(file)
	file.WriteString("{definitely not json}")
	internal.WriteBoundary(file)

	file.Close()

	att, err := OpenExe(file.Name())
	assert.EqualError(t, err, "corrupt attachment data (invalid TOC)")
	assert.Nil(t, att)
}

func TestOpenExe_offsetsToBig(t *testing.T) {
	var testAttachments = [][]byte{{1, 2, 3}}

	var testTOC = internal.TOC{
		internal.Attachment{
			Name: "att1",
			Size: 9000,
		},
	}

	path := prepareFile(t, testTOC, testAttachments)
	defer os.Remove(path)

	att, err := OpenExe(path)
	assert.EqualError(t, err, "corrupt attachment data (offsets too large)")
	assert.Nil(t, att)
}

func TestOpenExe_invalidOffsets(t *testing.T) {
	var testAttachments = [][]byte{{1, 2, 3}}

	var testTOC = internal.TOC{
		internal.Attachment{
			Name: "att1",
			Size: 2, // one byte less than expected
		},
	}

	path := prepareFile(t, testTOC, testAttachments)
	defer os.Remove(path)

	att, err := OpenExe(path)
	assert.EqualError(t, err, "corrupt attachment data (invalid offsets)")
	assert.Nil(t, att)
}
