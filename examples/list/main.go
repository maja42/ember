package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/maja42/ember"
)

func main() {
	attachments, err := ember.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer attachments.Close()

	fmt.Printf("Executable contains %d attachments\n", attachments.Count())
	contents := attachments.List()

	for _, name := range contents {
		s := attachments.Size(name)
		fmt.Printf("\nAttachment %q has %d bytes:\n", name, s)
		r := attachments.Reader(name)
		if _, err := io.Copy(os.Stdout, r); err != nil {
			log.Fatal(err)
		}
		fmt.Println()
	}
}
