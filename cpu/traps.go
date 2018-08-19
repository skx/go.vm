//
// This file contains the callbacks that the virtual machine
// can implement via the `int` operation.
//
//

package cpu

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

//
// TrapFunction is the signature for a function that is available
// as a trap.
//
type TrapFunction func(c *CPU, num int)

//
// TRAPS is an array of our trap-functions.
//
var TRAPS [0xffff]TrapFunction

//
// Helper for reading from stdin
//
var reader *bufio.Reader

//
// Trap Functions now follow
//

// TrapNOP is the default trap-function for any trap IDs that haven't
// explicitly been setup.
func TrapNOP(c *CPU, num int) {
	fmt.Printf("Trap function not defined: 0x%04X\n", num)
	os.Exit(1)
}

// StrLenTrap returns the length of a string.
//
// Input:
//   The string to measure in register 0.
// Output:
//   Sets register 0 with the length
//
func StrLenTrap(c *CPU, num int) {
	str := c.regs[0].GetString()
	c.regs[0].SetInt(len(str))
}

// ReadStringTrap reads a string from the console
//
// Input: None
//
// Ouptut:
//   Sets register 0 with the user-provided string
//
func ReadStringTrap(c *CPU, num int) {
	text, _ := reader.ReadString('\n')
	c.regs[0].SetString(text)
}

// RemoveNewLineTrap removes any trailing newline from the string in #0
//
// Input:
//   The string operate upon in #0.
// Output:
//   Sets register #0 with the updated string
//
func RemoveNewLineTrap(c *CPU, num int) {
	str := c.regs[0].GetString()
	c.regs[0].SetString(strings.TrimSpace(str))
}

// Now implement the traps
//
func init() {
	reader = bufio.NewReader(os.Stdin)
	TRAPS[0] = StrLenTrap
	TRAPS[1] = ReadStringTrap
	TRAPS[2] = RemoveNewLineTrap

	// Fill in the rest of the traps with
	// a function that will just report that
	// the given ID is not implemented
	for i := 0; i < 0xFFFF; i++ {
		if TRAPS[i] == nil {
			TRAPS[i] = TrapNOP
		}
	}
}
