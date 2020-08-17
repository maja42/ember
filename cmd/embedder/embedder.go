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

type CommandLine struct {
	Executable     string
	AttachmentList string
	Out            string
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

	fmt.Printf("Augmenting %q --> %q\n", cmd.Executable, cmd.Out)

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
	defer out.Close()

	err = embedding.EmbedFiles(out, exe, attachments, func(format string, args ...interface{}) {
		fmt.Printf("\t"+format+"\n", args...)
	})
	if err != nil {
		log.Fatalf("Failed to embed files: %s", err)
	}
	fmt.Printf("Finished")
}
