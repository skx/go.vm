package cpu

import (
	"bufio"
	"strings"
	"testing"

	"github.com/skx/go.vm/opcode"
)

// TestTraps will test the simple traps we have
func TestTraps(t *testing.T) {

	// Create CPU
	c := NewCPU()

	// Trap 0 -> Strlen
	c.LoadBytes([]byte{
		// Register 00 contains string "Steve\n"
		byte(opcode.STRING_STORE),
		00,

		// string length: 0006
		06,
		00,
		byte('S'), // "Steve\n"
		byte('t'),
		byte('e'),
		byte('v'),
		byte('e'),
		byte('\n'),

		// Trap 0 -> StrLenTrap
		byte(opcode.TRAP_OP),
		00,
		00,

		// Exit
		byte(opcode.EXIT),
	})

	err := c.Run()
	if err != nil {
		t.Fatalf("error running program: %s", err)
	}

	// Register 00 should have content 6
	val, err := c.regs[0].GetInt()
	if err != nil {
		t.Fatalf("error getting register contents")
	}
	if val != 6 {
		t.Fatalf("trap(strlen) test failed, got %d not 0x06", val)
	}

	//
	// Fake buffer for reading STDIN FROM.
	//
	strBuf := strings.NewReader("Hello, World!\n")
	c.STDIN = bufio.NewReader(strBuf)

	// Trap 1 -> ReadString
	c.LoadBytes([]byte{
		// Register 00 contains "0", via XOR
		byte(opcode.XOR_OP),
		00,
		00,
		00,

		// Trap 1 -> Read STDIN
		byte(opcode.TRAP_OP),
		01,
		00,

		// Exit
		byte(opcode.EXIT),
	})

	err = c.Run()
	if err != nil {
		t.Fatalf("error running program: %s", err)
	}

	// Register 00 should now contain a string.
	valS, errS := c.regs[0].GetString()
	if errS != nil {
		t.Fatalf("unexpected error getting stdin-value")
	}
	if valS != "Hello, World!\n" {
		t.Fatalf("Got wrong read from STDIN '%s'", valS)
	}

	// Trap 2 -> RemoveNewline
	c.LoadBytes([]byte{
		// Register 01 contains "Steve\n"
		// Register 00 contains string "Steve\n"
		byte(opcode.STRING_STORE),
		00,

		// string length: 0006
		06,
		00,
		byte('S'), // "Steve\n"
		byte('t'),
		byte('e'),
		byte('v'),
		byte('e'),
		byte('\n'),

		// Trap 1 -> Remove newline
		byte(opcode.TRAP_OP),
		02,
		00,

		// Trap 0 -> StrLenTrap
		byte(opcode.TRAP_OP),
		00,
		00,

		// Exit
		byte(opcode.EXIT),
	})

	err = c.Run()
	if err != nil {
		t.Fatalf("error running program: %s", err)
	}

	// Register 00 should have content 5
	val, err = c.regs[0].GetInt()
	if err != nil {
		t.Fatalf("error getting register contents")
	}
	if val != 5 {
		t.Fatalf("trap(strlen(stripnewline()) test failed, got %d not 0x05", val)
	}
}

// TestTrapError sets int not string in the appropriate register
func TestTrapError(t *testing.T) {

	// Create CPU
	c := NewCPU()

	for _, trap := range []int{0, 2} {
		// Trap 0 -> Strlen
		c.LoadBytes([]byte{
			// Register 00 contains "0", via XOR
			byte(opcode.XOR_OP),
			00,
			00,
			00,
			// Trap NN
			byte(opcode.TRAP_OP),
			byte(trap),
			00,

			// Exit
			byte(opcode.EXIT),
		})

		err := c.Run()
		if err == nil {
			t.Fatalf("expected error, got none")
		}
		if !strings.Contains(err.Error(), "attempting to call GetString") {

			t.Fatalf("expected error, got wrong one:%s", err)
		}
	}

}
