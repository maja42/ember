# List example

This is a minimalistic example example that prints the names, size and contents of all attachments.
 
## Building

First, build the example without embedding any attachments:

```
cd ./examples/list
go build
```

Running `./list` produces the following output:

```
Executable contains 0 attachments
```

## Embedding

Embedding happens at runtime after the target executable has already been built.
It is done via the embedder-cli. Build the executable and move it into
the list example directory:

```
go build -o examples/list/ ./cmd/embedder
```

Now you need some data to embed, as well as a json file describing the operation.

The example already contains an `attachments.json` that will embed the two files `fileA.txt` and `fileB.txt`.

Execute the following:

```
cd ./examples/list
./embedder -exe ./list -out ./newList
```

Output:

```
Augmenting "./list" --> "./newList"
        Adding TOC (57 bytes)
        Adding file A (33 bytes)
        Adding file B (33 bytes)
Finished
```

## Running 

Executing `./newList` now yields the following output:

```
Executable contains 2 attachments

Attachment "file A" has 33 bytes:
**This is the content of file A**

Attachment "file B" has 33 bytes:
**This is the content of file B**
```

## Cross-Platform

Embedding can be done on any platform and for any platform.
It is independent of the target-executable's format.

For example, it's possible to cross-compile `list` for windows, 
and add the attachments to the `.exe` on a Linux OS (and vice-versa):

```
cd ./examples/list
GOOS=windows GOARCH=amd64 go build
./embedder -exe ./list.exe -out ./newList.exe
```
