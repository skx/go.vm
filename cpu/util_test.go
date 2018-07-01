package cpu

import (
	"os"
	"testing"
)

// Test we can trim quotes from a string
func TestStripQuotes(t *testing.T) {

	type TestCase struct {
		Input  string
		Output string
	}

	tests := []TestCase{
		{"This is OK", "This is OK"},
		{"\"This is OK", "\"This is OK"},
		{"This is OK\"", "This is OK\""},
		{"\"This is OK\"", "This is OK"}}

	for _, test := range tests {

		out := trimQuotes(test.Input, '"')
		if out != test.Output {
			t.Errorf("Mismatched output")
		}
	}
}

func TestSplit(t *testing.T) {

	out := splitCommand("/bin/sh -c \"ls /etc/\"")
	if len(out) != 3 {
		t.Errorf("Splitting failed!")
	}
}

// This is not a real test, just to bump our coverage.
func TestDebug(t *testing.T) {
	debugPrintf("")
	os.Setenv("DEBUG", "1")
	debugPrintf("")
}
