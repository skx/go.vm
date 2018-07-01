// This file contains the implementation of the registers the
// CPU contains.

package cpu

import (
	"fmt"
	"os"
)

// Register holds the contents of a single register.
//
// This is horrid because we don't use an enum for the type.
type Register struct {
	// Integer contents of register if t == "int"
	i int

	// String contents of register if t == "string"
	s string

	// Register type: "int" vs. "string"
	t string
}

//
// Register functions
//
func NewRegister() *Register {
	r := &Register{}
	r.SetInt(0)
	return (r)
}

// GetInt retrieves the integer content of the given register.
// If the register contains a string that is a fatal error.
func (r *Register) GetInt() int {
	if r.t != "int" {
		fmt.Printf("BUG: Attempting to call GetInt on a register holding a non-integer value.\n")
		os.Exit(3)
	}
	return r.i
}

// SetInt stores the given integer in the register.
func (r *Register) SetInt(v int) {
	r.i = v
	r.t = "int"
}

// GetInt retrieves the string content of the given register.
// If the register contains an integer that is a fatal error.
func (r *Register) GetString() string {
	if r.t != "string" {
		fmt.Printf("BUG: Attempting to call GetString on a register holding a non-string value.\n")
		os.Exit(3)
	}
	return r.s
}

// SetString stores the supplied string in the register.
func (r *Register) SetString(v string) {
	r.s = v
	r.t = "string"
}

// Return the type of a registers contents `int` vs. `string`.
func (r *Register) Type() string {
	return (r.t)
}
