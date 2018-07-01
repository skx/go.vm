//
// This is a simple port of the virtual machine interpreter to golang.
//
// Ideally this would be broken into files for the structures, and
// the utility functions.
//
// We should probably also use `opcodes/opcodes.go` instead of the literal
// hex-constants for our bytecode.
//

package cpu

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"time"
)

// Flags holds the CPU flags - of which we only have one.
type Flags struct {
	// Zero-flag
	z bool
}

// CPU is our virtual machine state.
type CPU struct {
	// Registers
	regs [16]*Register

	// Flags
	flags Flags

	// Our RAM - where the program is loaded
	mem [0xFFFF]byte

	// Instruction-pointer
	ip int

	// stack
	stack *Stack
}

//
// CPU / VM functions
//

// NewCPU returns a new CPU object.
func NewCPU() *CPU {
	x := &CPU{}
	x.Reset()
	return x
}

// Reset sets the CPU into a known-good state, by setting the IP to zero,
// and emptying all registers (i.e. setting them to zero too).
func (c *CPU) Reset() {

	// Reset registers
	for i := 0; i < 16; i++ {
		c.regs[i] = NewRegister()
	}

	// Reset stack
	c.stack = NewStack()

	// Reset instruction pointer to zero.
	c.ip = 0
}

// LoadFile loads the program from the named file into RAM.
// NOTE: The CPU-state is reset prior to the load.
func (c *CPU) LoadFile(path string) {

	// Load the file
	b, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf("Failed to read file: %s - %s\n", path, err.Error())
		os.Exit(1)
	}

	if len(b) >= 0xFFFF {
		fmt.Printf("Program too large for RAM!\n")
		os.Exit(1)
	}

	// Copy contents of file to our memory region.
	// NOTE: This calls `Reset` too :)
	c.LoadBytes(b)
}

// LoadBytes populates the given program into RAM.
// NOTE: The CPU-state is reset prior to the load.
func (c *CPU) LoadBytes(data []byte) {

	// Ensure we reset our state.
	c.Reset()

	if len(data) >= 0xFFFF {
		fmt.Printf("Program too large for RAM!\n")
		os.Exit(1)
	}

	// Copy contents of file to our memory region
	for i := 0; i < len(data); i++ {
		c.mem[i] = data[i]
	}
}

// Read a string from the IP position
// Strings are prefixed by their lengths (two-bytes).
func (c *CPU) readString() string {
	// Read the length of the string we expect
	len := c.read2Val()

	// Now build up the body of the string
	s := ""
	for i := 0; i < len; i++ {
		s += string(c.mem[c.ip+i])
	}

	// Jump the IP over the length of the string.
	c.ip += (len)
	return s
}

// Read a two-byte number from the current IP.
// i.e This reads two bytes and returns a 16-bit value to the caller,
// skipping over both bytes in the IP.
func (c *CPU) read2Val() int {
	l := int(c.mem[c.ip])
	c.ip += 1
	h := int(c.mem[c.ip])
	c.ip += 1

	val := l + h*256
	return (val)
}

// Run launches our intepreter.
// It does not terminate until an `EXIT` instruction is hit.
func (c *CPU) Run() {
	run := true
	for run {

		instruction := c.mem[c.ip]
		debugPrintf("About to execute instruction %02X\n", instruction)

		switch instruction {
		case 0x00:
			debugPrintf("EXIT\n")
			run = false

		case 0x01:
			debugPrintf("INT_STORE\n")

			// register
			c.ip += 1
			reg := int(c.mem[c.ip])
			c.ip += 1
			val := c.read2Val()

			debugPrintf("\tSet register %02X to %04X\n", reg, val)
			c.regs[reg].SetInt(val)

		case 0x02:
			debugPrintf("INT_PRINT\n")
			// register
			c.ip += 1
			reg := c.mem[c.ip]

			val := c.regs[reg].GetInt()
			if val < 256 {
				fmt.Printf("%02X", val)
			} else {
				fmt.Printf("%04X", val)
			}
			c.ip += 1

		case 0x03:
			debugPrintf("INT_TOSTRING\n")
			// register
			c.ip += 1
			reg := c.mem[c.ip]

			// get value
			i := c.regs[reg].GetInt()

			// change from int to string
			c.regs[reg].SetString(fmt.Sprintf("%d", i))

			// next instruction
			c.ip += 1

		case 0x04:
			debugPrintf("INT_RANDOM\n")
			// register
			c.ip += 1
			reg := c.mem[c.ip]

			// New random source
			s1 := rand.NewSource(time.Now().UnixNano())
			r1 := rand.New(s1)

			// New random number
			c.regs[reg].SetInt(r1.Intn(0xffff))
			c.ip += 1

		case 0x10:
			debugPrintf("JUMP\n")
			c.ip += 1
			addr := c.read2Val()
			c.ip = addr

		case 0x11:
			debugPrintf("JUMP_Z\n")
			c.ip += 1
			addr := c.read2Val()
			if c.flags.z {
				c.ip = addr
			}

		case 0x12:
			debugPrintf("JUMP_NZ\n")
			c.ip += 1
			addr := c.read2Val()
			if !c.flags.z {
				c.ip = addr
			}

		case 0x20:
			debugPrintf("XOR\n")
			c.ip += 1
			res := c.mem[c.ip]
			c.ip += 1
			a := c.mem[c.ip]
			c.ip += 1
			b := c.mem[c.ip]
			c.ip += 1

			// store result
			a_val := c.regs[a].GetInt()
			b_val := c.regs[b].GetInt()
			c.regs[res].SetInt(a_val ^ b_val)

		case 0x21:
			debugPrintf("ADD\n")
			c.ip += 1
			res := c.mem[c.ip]
			c.ip += 1
			a := c.mem[c.ip]
			c.ip += 1
			b := c.mem[c.ip]
			c.ip += 1

			// store result
			a_val := c.regs[a].GetInt()
			b_val := c.regs[b].GetInt()
			c.regs[res].SetInt(a_val + b_val)

		case 0x22:
			debugPrintf("SUB\n")
			c.ip += 1
			res := c.mem[c.ip]
			c.ip += 1
			a := c.mem[c.ip]
			c.ip += 1
			b := c.mem[c.ip]
			c.ip += 1

			// store result
			a_val := c.regs[a].GetInt()
			b_val := c.regs[b].GetInt()
			c.regs[res].SetInt(a_val - b_val)

			// set the zero-flag if the result was zero or less
			if c.regs[res].i <= 0 {
				c.flags.z = true
			}

		case 0x23:
			debugPrintf("MUL\n")
			c.ip += 1
			res := c.mem[c.ip]
			c.ip += 1
			a := c.mem[c.ip]
			c.ip += 1
			b := c.mem[c.ip]
			c.ip += 1

			// store result
			a_val := c.regs[a].GetInt()
			b_val := c.regs[b].GetInt()
			c.regs[res].SetInt(a_val * b_val)

		case 0x24:
			debugPrintf("DIV\n")
			c.ip += 1
			res := c.mem[c.ip]
			c.ip += 1
			a := c.mem[c.ip]
			c.ip += 1
			b := c.mem[c.ip]
			c.ip += 1

			// store result
			a_val := c.regs[a].GetInt()
			b_val := c.regs[b].GetInt()

			if b_val == 0 {
				fmt.Printf("Attempting to divide by zero - denying\n")
				os.Exit(3)
			}
			c.regs[res].SetInt(a_val / b_val)

		case 0x25:
			debugPrintf("INC\n")

			// register
			c.ip += 1
			reg := c.mem[c.ip]

			// get the value
			val := c.regs[reg].GetInt()

			// if the value is the max it will wrap around
			if val == 0xFFFF {
				val = 0
			} else {
				// otherwise be incremented normally
				val += 1
			}

			// zero?
			c.flags.z = (val == 0)

			c.regs[reg].SetInt(val)

			// bump past that
			c.ip += 1

		case 0x26:
			debugPrintf("DEC\n")

			// register
			c.ip += 1
			reg := c.mem[c.ip]

			// get the value
			val := c.regs[reg].GetInt()

			// if the value is the minimum it will wrap around
			if val == 0x0000 {
				val = 0xFFFF
			} else {
				// otherwise decrease normally
				val -= 1
			}

			// zero?
			c.flags.z = (val == 0)

			c.regs[reg].SetInt(val)

			// bump past that
			c.ip += 1

		case 0x27:
			debugPrintf("AND\n")
			c.ip += 1
			res := c.mem[c.ip]
			c.ip += 1
			a := c.mem[c.ip]
			c.ip += 1
			b := c.mem[c.ip]
			c.ip += 1

			// store result
			a_val := c.regs[a].GetInt()
			b_val := c.regs[b].GetInt()
			c.regs[res].SetInt(a_val & b_val)

		case 0x28:
			debugPrintf("OR\n")
			c.ip += 1
			res := c.mem[c.ip]
			c.ip += 1
			a := c.mem[c.ip]
			c.ip += 1
			b := c.mem[c.ip]
			c.ip += 1

			// store result
			a_val := c.regs[a].GetInt()
			b_val := c.regs[b].GetInt()
			c.regs[res].SetInt(a_val | b_val)

		case 0x30:
			debugPrintf("STORE_STRING\n")

			// register
			c.ip += 1
			reg := c.mem[c.ip]

			// bump past that to the length + string
			c.ip += 1

			// read it
			str := c.readString()
			debugPrintf("\tRead String: '%s'\n", str)

			// store the string
			c.regs[reg].SetString(str)

		case 0x31:
			debugPrintf("PRINT_STRING\n")

			// register
			c.ip += 1
			reg := c.mem[c.ip]

			fmt.Printf("%s", c.regs[reg].GetString())
			c.ip += 1

		case 0x32:
			debugPrintf("STRING_CONCAT\n")

			// output register
			c.ip += 1
			res := c.mem[c.ip]

			// src1
			c.ip += 1
			a := c.mem[c.ip]

			// src2
			c.ip += 1
			b := c.mem[c.ip]

			c.ip += 1

			a_val := c.regs[a].GetString()
			b_val := c.regs[b].GetString()

			c.regs[res].SetString(a_val + b_val)

		case 0x33:
			debugPrintf("SYSTEM\n")

			// register
			c.ip += 1
			r := c.mem[c.ip]
			c.ip += 1

			// run the command
			toExec := splitCommand(c.regs[r].GetString())
			cmd := exec.Command(toExec[0], toExec[1:]...)

			var out bytes.Buffer
			var err bytes.Buffer
			cmd.Stdout = &out
			cmd.Stderr = &err
			cmd.Run()

			// stdout
			fmt.Printf("%s", out.String())

			// stderr - if non-empty
			if len(err.String()) > 0 {
				fmt.Printf("%s", err.String())
			}

		case 0x34:
			debugPrintf("STRING_TOINT\n")

			// register
			c.ip += 1
			reg := c.mem[c.ip]

			// get value
			s := c.regs[reg].GetString()

			i, err := strconv.Atoi(s)
			if err == nil {
				c.regs[reg].SetInt(i)
			} else {
				fmt.Printf("Failed to convert '%s' to int: %s", s, err.Error())
				os.Exit(3)
			}

			// next instruction
			c.ip += 1

		case 0x40:
			debugPrintf("CMP_REG\n")
			c.ip += 1
			r1 := int(c.mem[c.ip])
			c.ip += 1
			r2 := int(c.mem[c.ip])
			c.ip += 1

			c.flags.z = false

			switch c.regs[r1].Type() {
			case "int":
				if c.regs[r1].GetInt() == c.regs[r2].GetInt() {
					c.flags.z = true
				}
			case "string":
				if c.regs[r1].GetString() == c.regs[r2].GetString() {
					c.flags.z = true
				}
			}

		case 0x41:
			debugPrintf("CMP_IMMEDIATE\n")
			// register
			c.ip += 1
			reg := int(c.mem[c.ip])
			c.ip += 1
			val := c.read2Val()

			if c.regs[reg].Type() == "int" && c.regs[reg].GetInt() == val {
				c.flags.z = true
			} else {
				c.flags.z = false
			}

		case 0x42:
			debugPrintf("CMP_STR\n")
			// register
			c.ip += 1
			reg := int(c.mem[c.ip])
			c.ip += 1

			// read it
			str := c.readString()

			if c.regs[reg].Type() == "string" && c.regs[reg].GetString() == str {
				c.flags.z = true
			} else {
				c.flags.z = false
			}

		case 0x43:
			debugPrintf("IS_STRING\n")
			// register
			c.ip += 1
			reg := int(c.mem[c.ip])
			c.ip += 1

			if c.regs[reg].Type() == "string" {
				c.flags.z = true
			} else {
				c.flags.z = false
			}

		case 0x44:
			debugPrintf("IS_INT\n")
			// register
			c.ip += 1
			reg := int(c.mem[c.ip])
			c.ip += 1

			if c.regs[reg].Type() == "int" {
				c.flags.z = true
			} else {
				c.flags.z = false
			}

		case 0x50:
			debugPrintf("NOP\n")
			c.ip += 1

		case 0x51:
			debugPrintf("STORE\n")

			// register
			c.ip += 1
			dst := int(c.mem[c.ip])
			c.ip += 1

			// register
			src := int(c.mem[c.ip])
			c.ip += 1

			c.regs[src] = c.regs[dst]

		case 0x60:
			debugPrintf("PEEK\n")

			// register
			c.ip += 1
			result := int(c.mem[c.ip])

			c.ip += 1
			src := int(c.mem[c.ip])

			// get the address from the src register contents
			addr := c.regs[src].i

			// store the contents of the given address
			c.regs[result].SetInt(int(c.mem[addr]))
			c.ip += 1

		case 0x61:
			debugPrintf("POKE\n")

			// register
			c.ip += 1
			src := int(c.mem[c.ip])
			c.ip += 1

			dst := int(c.mem[c.ip])
			c.ip += 1

			// So the destination will contain an address
			// put the contents of the source to that.
			addr := c.regs[dst].i
			val := c.regs[src].i

			debugPrintf("Writing %02X to %04X\n", val, addr)
			c.mem[addr] = byte(val)

		case 0x62:
			debugPrintf("MEMCPY\n")

			// register
			c.ip += 1
			dst := int(c.mem[c.ip])
			c.ip += 1

			src := int(c.mem[c.ip])
			c.ip += 1

			len := int(c.mem[c.ip])
			c.ip += 1

			// get the addresses from the registers
			src_addr := c.regs[src].GetInt()
			dst_addr := c.regs[dst].GetInt()
			length := c.regs[len].GetInt()

			i := 0
			for i < length {

				if dst_addr >= 0xFFFF {
					dst_addr = 0
				}
				if src_addr >= 0xFFFF {
					src_addr = 0
				}

				c.mem[dst_addr] = c.mem[src_addr]
				dst_addr += 1
				src_addr += 1
				i += 1
			}

		case 0x70:
			debugPrintf("PUSH\n")

			// register
			c.ip += 1
			reg := int(c.mem[c.ip])
			c.ip += 1

			// Store the value in the register on the stack
			c.stack.Push(c.regs[reg].GetInt())

		case 0x71:
			debugPrintf("POP\n")

			// register
			c.ip += 1
			reg := int(c.mem[c.ip])
			c.ip += 1

			// Ensure our stack isn't empty
			if c.stack.Empty() {
				fmt.Printf("Stack Underflow!\n")
				os.Exit(1)
			}
			// Store the value in the register on the stack
			val, _ := c.stack.Pop()
			c.regs[reg].SetInt(val)

		case 0x72:
			debugPrintf("RET\n")

			// Ensure our stack isn't empty
			if c.stack.Empty() {
				fmt.Printf("Stack Underflow!\n")
				os.Exit(1)
			}

			// Get the address
			addr, _ := c.stack.Pop()

			// jump
			c.ip = addr

		case 0x73:
			debugPrintf("CALL\n")
			c.ip += 1

			addr := c.read2Val()

			// push the current IP onto the stack
			c.stack.Push(c.ip)

			// jump to the call address
			c.ip = addr

		case 0x80:
			debugPrintf("TRAP\n")
			c.ip += 1

			num := c.read2Val()

			fn := TRAPS[num]
			if fn != nil {
				fn(c, num)
			}
		default:
			fmt.Printf("Unrecognized/Unimplemented opcode %02X at IP %04X\n", instruction, c.ip)
			os.Exit(1)
		}

		// Ensure our instruction-pointer wraps around.
		if c.ip > 0xFFFF {
			c.ip = 0
		}
	}
}
