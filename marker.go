package ember

import (
	"fmt"
	"time"
)

// marker is compiled into executables which accept attachments.
// This allows the augmenter to verify that the target file is compatible.
var marker = "~~MagicMarker for maja42/ember/v1~~"

func init() {
	// Dead code that uses 'marker' and is not eliminated by the compiler.
	if time.Now().Nanosecond() == -42 {
		fmt.Print(marker)
	}
}
