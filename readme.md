# ember  [![GoDoc](https://godoc.org/github.com/maja42/ember?status.svg)](https://godoc.org/github.com/maja42/ember)

Ember is a tool for embedding arbitrary resources into a go executable at runtime.
The difference to conventional embedding libraries is that the resources don't need to exist at build-time.

Embedding binary files (eg. zip-archives) is supported.

## Use case

Often, executables require runtime- or user-defined configurations and resources that need to be stored alongside
the executable.
This forces the end-user to deal with multiple files when copying/moving/distributing the application
and also makes it possible to manipulate said configuration, which is not always desirable.

The main use-case of ember  is to bundle such configuration files and other resources with the application at runtime.
There is no need for providing a go-toolchain and re-building the application each time there is a new configuration.

## Cross platform

Ember is truly cross-platform. It does not only support any OS, but embedding resources can also be done cross-platform.

## Example

examples/list contains a minimalistic example application using ember.

## Usage

The following example shows how to access embedded files from within the application:

```go
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
		io.Copy(os.Stdout, r)
		fmt.Println()
	}
}
```

To embed files in the first place, create an `attachments.json` file describing the files to embed:

```json
{
  "file A": "path/to/file A.txt",
  "file B": "path/to/file B.txt"
}
```

Afterwards, use cmd/embedder to attach the files to an already-built executable:

```bash
./embedder -exe app -attachments attachments.json -out finishedApp
```

The full workflow is shown in examples/list.

## Contributions

Feel free to submit feature requests, bug reports and pull requests.
