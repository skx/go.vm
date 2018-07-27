// Some utility functions are bundled here - they're primarily used by
// the `system` opcode.

package cpu

import (
	"fmt"
	"os"
	"regexp"
)

//
// Global functions
//

// debugPrintf outputs some debugging details when `$DEBUG=1`.
func debugPrintf(format string, args ...interface{}) {
	if os.Getenv("DEBUG") == "" {
		return
	}
	prefix := fmt.Sprintf("%s", format)
	fmt.Printf(prefix, args...)
}

// Split a line of text into tokens, but keep anything "quoted"
// together.
//
// So this input:
//
//   /bin/sh -c "ls /etc"
//
// Would give output of the form:
//   /bin/sh
//   -c
//   ls /etc
//
func splitCommand(input string) []string {

	//
	// This does the split into an array
	//
	r := regexp.MustCompile(`[^\s"']+|"([^"]*)"|'([^']*)`)
	res := r.FindAllString(input, -1)

	//
	// However the resulting pieces might be quoted.
	// So we have to remove them, if present.
	//
	var result []string
	for _, e := range res {
		result = append(result, trimQuotes(e, '"'))
	}
	return (result)
}

// Remove balanced characters around a string.
func trimQuotes(in string, c byte) string {
	if len(in) >= 2 {
		if in[0] == c && in[len(in)-1] == c {
			return in[1 : len(in)-1]
		}
	}
	return in
}
