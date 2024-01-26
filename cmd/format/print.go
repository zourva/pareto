package format

import (
	"fmt"
	"os"
)

func PrintPerLine(showNumber bool, args ...string) {
	for i, arg := range args {
		if showNumber {
			_, _ = fmt.Fprintln(os.Stderr, i+1, arg)
		} else {
			_, _ = fmt.Fprintln(os.Stderr, arg)
		}
	}
}

func PrintOneLine(args ...string) {
	for _, arg := range args {
		_, _ = fmt.Fprint(os.Stderr, arg, " ")
	}

	_, _ = fmt.Fprintln(os.Stderr)
}
