//go:build go1.18
// +build go1.18

package cpu

import (
	"context"
	"strings"
	"testing"
	"time"
)

// FuzzEvaluator runs the fuzz-testing against our evaluation engine
func FuzzCPU(f *testing.F) {

	// empty + whitespace
	f.Add([]byte(nil))
	f.Add([]byte(""))

	// Known errors we might see
	known := []string{
		"Unimplemented opcode",
		"attempted division by zero",
		"attempting to call GetInt on a register holding a",
		"attempting to call GetString on a register holding a",
		"error invoking system",
		"out of range",
		"stackunderflow",
		"strconv",
		"timeout during execution",
		"too large",
		"Trap function not defined:",
		"invalid trap ",
	}
	f.Fuzz(func(t *testing.T, input []byte) {

		// Create CPU
		c := NewCPU()

		// Avoid terminating on infinite loops
		ctx, cancel := context.WithTimeout(
			context.Background(),
			500*time.Millisecond,
		)
		defer cancel()

		// Set the context
		c.SetContext(ctx)

		// Load the program
		c.LoadBytes(input)

		// Run it.
		err := c.Run()

		// Found an error
		if err != nil {

			falsePositive := false

			// is it a known one?
			for _, expected := range known {
				if strings.Contains(err.Error(), expected) {
					falsePositive = true
				}
			}

			if !falsePositive {
				t.Fatalf("error running input %s -> %s", input, err.Error())
			}

		}
	})
}
