package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/maja42/ember/embedding"
)

// CommandLine configuration
type CommandLine struct {
	Executable      string
	RemoveEmbedding bool
	AttachmentList  string
	Out             string
}

// AttachmentList maps embedded files (arbitrary name) to paths where they can be found on the filesystem.
type AttachmentList map[string]string

// LoadAttachmentList loads the list of attachments from a json file.
func LoadAttachmentList(path string) AttachmentList {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		log.Panicf("Failed to open attachment list %q: %s", path, err)
	}

	var list AttachmentList
	if err := json.Unmarshal(file, &list); err != nil {
		log.Panicf("Failed to read attachment list: %s", err)
	}
	return list
}

func main() {
	var cmd CommandLine
	flag.StringVar(&cmd.Executable, "exe", "", "Target executable that should be modified (windows or linux)")
	flag.BoolVar(&cmd.RemoveEmbedding, "remove", false, "If attachments should be removed from an already augmented executable")
	flag.StringVar(&cmd.AttachmentList, "attachments", "attachments.json", "Path to JSON file containing a list of attachments to embed")
	flag.StringVar(&cmd.Out, "out", "", "Path for the resulting executable")
	flag.Parse()
	if cmd.Executable == "" || cmd.Out == "" {
		flag.Usage()
		os.Exit(1)
	}
	if !cmd.RemoveEmbedding && cmd.AttachmentList == "" { // nothing to do?
		flag.Usage()
		os.Exit(1)
	}

	// Open executable
	exe, err := os.Open(cmd.Executable)
	if err != nil {
		log.Fatalf("Failed to open executable %q: %s", cmd.Executable, err)
	}
	defer exe.Close()

	// Open output
	out, err := os.OpenFile(cmd.Out, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0755)
	if err != nil {
		log.Fatalf("Failed to open output file %q: %s", cmd.Out, err)
	}
	defer func() {
		_ = out.Close()
		if err := recover(); err != nil { // execution failed; delete created output file
			_ = os.Remove(cmd.Out)
		}
	}()

	logger := func(format string, args ...interface{}) {
		fmt.Printf("\t"+format+"\n", args...)
	}

	if cmd.RemoveEmbedding {
		fmt.Printf("Removing embedded content from %q --> %q", cmd.Executable, cmd.Out)

		err = embedding.RemoveEmbedding(out, exe, logger)
		if err != nil {
			log.Panicf("Failed to remove embedded content: %s", err)
		}
	} else {
		fmt.Printf("Augmenting %q --> %q", cmd.Executable, cmd.Out)

		attachments := LoadAttachmentList(cmd.AttachmentList)
		err = embedding.EmbedFiles(out, exe, attachments, logger)
		if err != nil {
			log.Panicf("Failed to embed files: %s", err)
		}
	}

	fmt.Printf("Finished")
}
