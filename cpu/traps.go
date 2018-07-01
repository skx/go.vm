//
// This file contains the callbacks that the virtual machine
// can implement via the `int` operation.
//
//

package cpu

import (
	"bufio"
	"os"
)

//
// TrapFunction is the signature for a function that is available
// as a trap.
//
type TrapFunction func(c *CPU)

//
// Create an array of trap-functions.
//
var TRAPS [0xffff]TrapFunction

//
// Helper for reading from stdin
//
var reader *bufio.Reader

//
// Trap Functions now follow
//

// StrLenTrap returns the length of a string.
//
// Input:
//   The string to measure in register 0.
// Output:
//   Sets register 0 with the length
//
func StrLenTrap(c *CPU) {
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
func ReadStringTrap(c *CPU) {
	text, _ := reader.ReadString('\n')
	c.regs[0].SetString(text)
}

// Now implement the traps
//
func init() {
	reader = bufio.NewReader(os.Stdin)
	TRAPS[0] = StrLenTrap
	TRAPS[1] = ReadStringTrap
}
