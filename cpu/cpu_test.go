package cpu

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/skx/go.vm/opcode"
)

// TestTimeout tests that we can catch infinite loops.
func TestTimeout(t *testing.T) {

	// Timeout after 0.150 seconds
	ctx, cancel := context.WithTimeout(
		context.Background(),
		150*time.Millisecond,
	)
	defer cancel()

	// Create the CPU
	c := NewCPU()

	// Setup the timeout
	c.SetContext(ctx)

	// Load the program
	c.LoadBytes([]byte{
		// Infinite loop
		byte(opcode.JUMP_TO),
		00,
		00,
	})

	// Run it
	err := c.Run()
	if err == nil {
		t.Fatalf("expected an error running program, got none")
	}
	if !strings.Contains(err.Error(), "timeout") {
		t.Fatalf("got an error, but the wrong one: %s", err)
	}

}

// TestDec will test that the decrement instruction does the right thing
func TestDec(t *testing.T) {

	c := NewCPU()
	c.LoadBytes([]byte{
		// Register 01 contains "0", via XOR
		byte(opcode.XOR_OP),
		01,
		01,
		01,

		// DEC register 01
		byte(opcode.DEC_OP),
		01,
		byte(opcode.DEC_OP),
		01,

		// Register 6 contains 33
		byte(opcode.INT_STORE),
		6,
		33,
		0,

		byte(opcode.DEC_OP),
		06,

		// Exit
		byte(opcode.EXIT),
	})

	err := c.Run()
	if err != nil {
		t.Fatalf("error running program")
	}

	// Register 01 should have content 254. (0 -1 - 1)
	val, err := c.regs[1].GetInt()
	if err != nil {
		t.Fatalf("error getting register contents")
	}
	if val != 65534 {
		t.Fatalf("inc test failed, got %d not 65534", val)
	}

	// register 06 should have content 32 (33 - 1)
	val, err = c.regs[6].GetInt()
	if err != nil {
		t.Fatalf("error getting register contents")
	}
	if val != 32 {
		t.Fatalf("inc test failed, got %d not 34", val)
	}
}

// TestInc will test that the increment instruction does the right thing
func TestInc(t *testing.T) {

	c := NewCPU()
	c.LoadBytes([]byte{
		// Register 01 contains "0", via XOR
		byte(opcode.XOR_OP),
		01,
		01,
		01,

		// INC register 01
		byte(opcode.INC_OP),
		01,

		// Register 6 contains 33
		byte(opcode.INT_STORE),
		6,
		33,
		0,

		byte(opcode.INC_OP),
		06,

		// Exit
		byte(opcode.EXIT),
	})

	err := c.Run()
	if err != nil {
		t.Fatalf("error running program")
	}

	// Register 01 should have content 1. (0 + 1)
	val, err := c.regs[1].GetInt()
	if err != nil {
		t.Fatalf("error getting register contents")
	}
	if val != 1 {
		t.Fatalf("inc test failed, got %d not 0x01", val)
	}

	// register 06 should have content 34 (33+ 1)
	val, err = c.regs[6].GetInt()
	if err != nil {
		t.Fatalf("error getting register contents")
	}
	if val != 34 {
		t.Fatalf("inc test failed, got %d not 34", val)
	}
}

func TestString2Int(t *testing.T) {

	c := NewCPU()
	c.LoadBytes([]byte{
		// Register 01 contains string
		byte(opcode.STRING_STORE),
		01,

		// string length: 0005
		05,
		00,
		byte('1'), // "12345"
		byte('2'),
		byte('3'),
		byte('4'),
		byte('5'),

		// Convert to int
		byte(opcode.STRING_TOINT),
		01,

		// Exit
		byte(opcode.EXIT),
	})

	err := c.Run()
	if err != nil {
		t.Fatalf("error running program")
	}

	val, err := c.regs[1].GetInt()
	if err != nil {
		t.Fatalf("error getting register contents as int:%s", err)
	}
	if val != 12345 {
		t.Fatalf("converted string to int failed: %d", val)
	}
}

func TestInt2String(t *testing.T) {

	c := NewCPU()
	c.LoadBytes([]byte{
		// Register 01 contains number 0xffff
		byte(opcode.INT_STORE),
		01,
		0xff,
		0xff,

		// Convert to int
		byte(opcode.INT_TOSTRING),
		01,

		// Exit
		byte(opcode.EXIT),
	})

	err := c.Run()
	if err != nil {
		t.Fatalf("error running program")
	}

	// Register 01 should have a string
	val, err := c.regs[1].GetString()
	if err != nil {
		t.Fatalf("error getting register contents as string:%s", err)
	}
	if val != "65535" {
		t.Fatalf("converted int to string failed: %s", val)
	}
}

// TestRegisterBounds tests using operations with opcodes representing registers
// that are out of bounds.
func TestRegisterBounds(t *testing.T) {

	type TestCase struct {
		Program []byte
	}

	tests := []TestCase{
		TestCase{
			// INT_STORE
			Program: []byte{
				byte(opcode.INT_STORE),
				0100,
				0xff,
				0xff,
			},
		},
		TestCase{
			// INT_PRINT
			Program: []byte{
				byte(opcode.INT_PRINT),
				0100,
				0xff,
				0xff,
			},
		},
		TestCase{
			// INT_TOSTRING
			Program: []byte{
				byte(opcode.INT_TOSTRING),
				0100,
				0xff,
				0xff,
			},
		},
		TestCase{
			// INT_RANDOM
			Program: []byte{
				byte(opcode.INT_RANDOM),
				0100,
			},
		},
		TestCase{
			// XOR
			Program: []byte{
				byte(opcode.XOR_OP),
				0xff,
				0x02,
				0x03,
			},
		},
		TestCase{
			Program: []byte{
				byte(opcode.XOR_OP),
				0x01,
				0xFF,
				0x02,
			},
		},
		TestCase{
			Program: []byte{
				byte(opcode.XOR_OP),
				0x01,
				0x02,
				0xff,
			},
		},
		TestCase{
			// ADD
			Program: []byte{
				byte(opcode.ADD_OP),
				0xff,
				0x02,
				0x03,
			},
		},
		TestCase{
			Program: []byte{
				byte(opcode.ADD_OP),
				0x01,
				0xFF,
				0x02,
			},
		},
		TestCase{
			Program: []byte{
				byte(opcode.ADD_OP),
				0x01,
				0x02,
				0xff,
			},
		},
		TestCase{
			// SUB
			Program: []byte{
				byte(opcode.SUB_OP),
				0xff,
				0x02,
				0x03,
			},
		},
		TestCase{
			Program: []byte{
				byte(opcode.SUB_OP),
				0x01,
				0xFF,
				0x02,
			},
		},
		TestCase{
			Program: []byte{
				byte(opcode.SUB_OP),
				0x01,
				0x02,
				0xff,
			},
		},
		TestCase{
			// MUL
			Program: []byte{
				byte(opcode.MUL_OP),
				0xff,
				0x02,
				0x03,
			},
		},
		TestCase{
			Program: []byte{
				byte(opcode.MUL_OP),
				0x01,
				0xFF,
				0x02,
			},
		},
		TestCase{
			Program: []byte{
				byte(opcode.MUL_OP),
				0x01,
				0x02,
				0xff,
			},
		},
		TestCase{
			// DIV
			Program: []byte{
				byte(opcode.DIV_OP),
				0xff,
				0x02,
				0x03,
			},
		},
		TestCase{
			Program: []byte{
				byte(opcode.DIV_OP),
				0x01,
				0xFF,
				0x02,
			},
		},
		TestCase{
			Program: []byte{
				byte(opcode.DIV_OP),
				0x01,
				0x02,
				0xff,
			},
		},
		TestCase{
			Program: []byte{
				byte(opcode.DIV_OP),
				0x01,
				0x02,
				0xff,
			},
		},
		TestCase{
			// INC
			Program: []byte{
				byte(opcode.INC_OP),
				0xff,
			},
		},
		TestCase{
			// DEC
			Program: []byte{
				byte(opcode.DEC_OP),
				0xff,
			},
		},
		TestCase{
			// AND
			Program: []byte{
				byte(opcode.AND_OP),
				0xff,
				0x02,
				0x03,
			},
		},
		TestCase{
			Program: []byte{
				byte(opcode.AND_OP),
				0x01,
				0xFF,
				0x02,
			},
		},
		TestCase{
			Program: []byte{
				byte(opcode.AND_OP),
				0x01,
				0x02,
				0xff,
			},
		},
		TestCase{
			// OR
			Program: []byte{
				byte(opcode.OR_OP),
				0xff,
				0x02,
				0x03,
			},
		},
		TestCase{
			Program: []byte{
				byte(opcode.OR_OP),
				0x01,
				0xFF,
				0x02,
			},
		},
		TestCase{
			Program: []byte{
				byte(opcode.OR_OP),
				0x01,
				0x02,
				0xff,
			},
		},
		TestCase{
			// STRING_STORE
			Program: []byte{
				byte(opcode.STRING_STORE),
				0xff,
			},
		},
		TestCase{
			// STRING_PRINT
			Program: []byte{
				byte(opcode.STRING_PRINT),
				0xff,
			},
		},
		TestCase{
			// STRING_CONCAT
			Program: []byte{
				byte(opcode.STRING_CONCAT),
				0xff,
				0x02,
				0x03,
			},
		},
		TestCase{
			Program: []byte{
				byte(opcode.STRING_CONCAT),
				0x01,
				0xFF,
				0x02,
			},
		},
		TestCase{
			Program: []byte{
				byte(opcode.STRING_CONCAT),
				0x01,
				0x02,
				0xff,
			},
		},
		TestCase{
			// STRING_SYSTEM
			Program: []byte{
				byte(opcode.STRING_SYSTEM),
				0xff,
			},
		},
		TestCase{
			// STRING_TOINT
			Program: []byte{
				byte(opcode.STRING_TOINT),
				0xff,
			},
		},
		TestCase{
			// CMP_REG
			Program: []byte{
				byte(opcode.CMP_REG),
				0xff,
				0x02,
			},
		},
		TestCase{
			Program: []byte{
				byte(opcode.CMP_REG),
				0x01,
				0xff,
			},
		},
		TestCase{
			// CMP_IMMEDIATE
			Program: []byte{
				byte(opcode.CMP_IMMEDIATE),
				0xff,
			},
		},
		TestCase{
			// CMP_STRING
			Program: []byte{
				byte(opcode.CMP_STRING),
				0xff,
			},
		},
		TestCase{
			// IS_STRING
			Program: []byte{
				byte(opcode.IS_STRING),
				0xff,
			},
		},
		TestCase{
			// IS_INTEGER
			Program: []byte{
				byte(opcode.IS_INTEGER),
				0xff,
			},
		},
		TestCase{
			// REG_STORE
			Program: []byte{
				byte(opcode.REG_STORE),
				0xff,
				0x02,
			},
		},
		TestCase{
			// REG_STORE
			Program: []byte{
				byte(opcode.REG_STORE),
				0x01,
				0xff,
			},
		},
		TestCase{
			// PEEK
			Program: []byte{
				byte(opcode.PEEK),
				0x01,
				0xff,
			},
		},
		TestCase{
			Program: []byte{
				byte(opcode.PEEK),
				0x01,
				0xff,
			},
		},
		TestCase{
			// POKE
			Program: []byte{
				byte(opcode.POKE),
				0x01,
				0xff,
			},
		},
		TestCase{
			Program: []byte{
				byte(opcode.POKE),
				0x01,
				0xff,
			},
		},
		TestCase{
			// MEMCPY
			Program: []byte{
				byte(opcode.MEMCPY),
				0xff,
				0x02,
				0x03,
			},
		},
		TestCase{
			Program: []byte{
				byte(opcode.MEMCPY),
				0x01,
				0xff,
				0x03,
			},
		},
		TestCase{
			Program: []byte{
				byte(opcode.MEMCPY),
				0x01,
				0x02,
				0xff,
			},
		},
		TestCase{
			// STACK_PUSH
			Program: []byte{
				byte(opcode.STACK_PUSH),
				0xff,
			},
		},
		TestCase{
			// STACK_POP
			Program: []byte{
				byte(opcode.STACK_POP),
				0xff,
			},
		},
	}

	for _, test := range tests {
		c := NewCPU()
		c.LoadBytes(test.Program)

		err := c.Run()
		if err == nil {
			t.Fatalf("expected and error running program, got none")
		}
		if !strings.Contains(err.Error(), "out of range") {
			t.Fatalf("got an error, but the wrong one: %s", err.Error())
		}

	}
}
