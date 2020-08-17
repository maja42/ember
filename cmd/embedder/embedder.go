package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/maja42/ember/internal"
)

type CommandLine struct {
	Executable     string
	AttachmentList string
	Out            string
}

const compatibleVersion = "maja42/ember/v1"

// ValidateExe ensures that the executable-to-modify is compatible with this embedder.
func ValidateExe(path string) io.ReadCloser {
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("Failed to open unzipper %q: %s", path, err)
	}

	// Check if the target executable is compatible.
	// Compatible executables are importing 'ember' in the correct version,
	// causing a marker-string to be present in the binary.
	// String-replace is used to ensure the marker is not present in the embedder-executable.
	marker := "~~MagicMarker for XXX~~"
	marker = strings.ReplaceAll(marker, "XXX", compatibleVersion)

	offset := internal.SeekPattern(file, []byte(marker))
	if offset == -1 { // does not import correct library(-version)
		log.Fatalf("Incompatible executable (magic string not found)")
	}

	offset = internal.SeekBoundary(file)
	if offset != -1 {
		log.Fatalf("Executable already contains embedded content")
	}

	if _, err = file.Seek(0, io.SeekStart); err != nil {
		log.Fatalf("Failed to seek: %s", err)
	}
	return file
}

type AttachmentList map[string]string

func LoadAttachmentList(path string) AttachmentList {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Failed to open attachment list %q: %s", path, err)
	}

	var list AttachmentList
	if err := json.Unmarshal(file, &list); err != nil {
		log.Fatalf("Failed to read attachment list: %s", err)
	}
	return list
}

func BuildTOC(attachments AttachmentList) internal.TOC {
	toc := make(internal.TOC, 0, len(attachments))

	for name, path := range attachments {
		info, err := os.Stat(path)
		if err != nil {
			log.Fatalf("Failed to probe attachment %q at %q: %s", name, path, err)
		}
		if info.IsDir() {
			log.Fatalf("Cannot attach directories: %q at %q", name, path)
		}
		toc = append(toc, internal.Attachment{
			Name: name,
			Size: info.Size(),
		})
	}
	return toc
}

func main() {
	var cmd CommandLine
	flag.StringVar(&cmd.Executable, "exe", "", "Target executable that should be modified (windows or linux)")
	flag.StringVar(&cmd.AttachmentList, "attachments", "attachments.json", "Path to JSON file containing a list of attachments to embed")
	flag.StringVar(&cmd.Out, "out", "", "Path for the resulting executable")
	flag.Parse()
	if cmd.Executable == "" || cmd.AttachmentList == "" || cmd.Out == "" {
		flag.Usage()
		os.Exit(1)
	}

	attachments := LoadAttachmentList(cmd.AttachmentList)
	toc := BuildTOC(attachments)

	exe := ValidateExe(cmd.Executable)
	defer exe.Close()

	fmt.Printf("Augmenting %q --> %q\n", cmd.Executable, cmd.Out)

	// Open output
	out, err := os.OpenFile(cmd.Out, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0755)
	if err != nil {
		log.Fatalf("Failed to open output file %q: %s", cmd.Out, err)
	}
	defer out.Close()

	// Write original executable
	if _, err := io.Copy(out, exe); err != nil {
		log.Fatalf("Failed to write output: %s", err)
	}
	// Boundary
	if err := internal.WriteBoundary(out); err != nil {
		log.Fatalf("Failed to write output: %s", err)
	}

	// TOC
	jsonTOC, err := json.Marshal(toc)
	if err != nil {
		log.Fatalf("Failed to marshal table of contents: %s", err)
	}

	fmt.Printf("\tAdding TOC (%d bytes)\n", len(jsonTOC))

	if _, err := out.Write(jsonTOC); err != nil {
		log.Fatalf("Failed to write output: %s", err)
	}
	// Boundary
	if err := internal.WriteBoundary(out); err != nil {
		log.Fatalf("Failed to write output: %s", err)
	}
	// Attachments
	for _, att := range toc {
		fmt.Printf("\tAdding %s (%d bytes)\n", att.Name, att.Size)

		path := attachments[att.Name]
		attachment, err := os.Open(path)
		if err != nil {
			log.Fatalf("Failed to open attachment %q at %q: %s", att.Name, path, err)
		}
		if _, err := io.Copy(out, attachment); err != nil {
			log.Fatalf("Failed to write output: %s", err)
		}
		_ = attachment.Close()
	}
	// Boundary
	if err := internal.WriteBoundary(out); err != nil {
		log.Fatalf("Failed to write output: %s", err)
	}

	fmt.Printf("Finished")
}
