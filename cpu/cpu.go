//
// This is a simple port of the virtual machine interpreter to golang.
//
// We should probably use the constants defined in `opcodes/opcodes.go`
// instead of the literal hex-constants for our bytecode, but that's a minor
// issue.
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

	"github.com/skx/go.vm/opcode"
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

		op := opcode.NewOpcode(c.mem[c.ip])
		debugPrintf("%04X %02X [%s]\n", c.ip, op.Value(), op.String())

		switch int(op.Value()) {
		case opcode.EXIT:
			run = false

		case opcode.INT_STORE:
			// register
			c.ip += 1
			reg := int(c.mem[c.ip])
			if reg < 0 || reg > 15 {
				fmt.Printf("Register %d out of range\n", reg)
				os.Exit(1)
			}

			c.ip += 1
			val := c.read2Val()
			c.regs[reg].SetInt(val)

		case opcode.INT_PRINT:
			// register
			c.ip += 1
			reg := c.mem[c.ip]

			if reg < 0 || reg > 15 {
				fmt.Printf("Register %d out of range\n", reg)
				os.Exit(1)
			}
			val := c.regs[reg].GetInt()
			if val < 256 {
				fmt.Printf("%02X", val)
			} else {
				fmt.Printf("%04X", val)
			}
			c.ip += 1

		case opcode.INT_TOSTRING:
			// register
			c.ip += 1
			reg := c.mem[c.ip]

			if reg < 0 || reg > 15 {
				fmt.Printf("Register %d out of range\n", reg)
				os.Exit(1)
			}

			// get value
			i := c.regs[reg].GetInt()

			// change from int to string
			c.regs[reg].SetString(fmt.Sprintf("%d", i))

			// next instruction
			c.ip += 1

		case opcode.INT_RANDOM:
			// register
			c.ip += 1
			reg := c.mem[c.ip]

			if reg < 0 || reg > 15 {
				fmt.Printf("Register %d out of range\n", reg)
				os.Exit(1)
			}

			// New random source
			s1 := rand.NewSource(time.Now().UnixNano())
			r1 := rand.New(s1)

			// New random number
			c.regs[reg].SetInt(r1.Intn(0xffff))
			c.ip += 1

		case opcode.JUMP_TO:
			c.ip += 1
			addr := c.read2Val()
			c.ip = addr

		case opcode.JUMP_Z:
			c.ip += 1
			addr := c.read2Val()
			if c.flags.z {
				c.ip = addr
			}

		case opcode.JUMP_NZ:
			c.ip += 1
			addr := c.read2Val()
			if !c.flags.z {
				c.ip = addr
			}

		case opcode.XOR_OP:
			c.ip += 1
			res := c.mem[c.ip]
			c.ip += 1
			a := c.mem[c.ip]
			c.ip += 1
			b := c.mem[c.ip]
			c.ip += 1

			// store result
			aVal := c.regs[a].GetInt()
			bVal := c.regs[b].GetInt()
			c.regs[res].SetInt(aVal ^ bVal)

		case opcode.ADD_OP:
			c.ip += 1
			res := c.mem[c.ip]
			c.ip += 1
			a := c.mem[c.ip]
			c.ip += 1
			b := c.mem[c.ip]
			c.ip += 1

			// store result
			aVal := c.regs[a].GetInt()
			bVal := c.regs[b].GetInt()
			c.regs[res].SetInt(aVal + bVal)

		case opcode.SUB_OP:
			c.ip += 1
			res := c.mem[c.ip]
			c.ip += 1
			a := c.mem[c.ip]
			c.ip += 1
			b := c.mem[c.ip]
			c.ip += 1

			// store result
			aVal := c.regs[a].GetInt()
			bVal := c.regs[b].GetInt()
			c.regs[res].SetInt(aVal - bVal)

			// set the zero-flag if the result was zero or less
			if c.regs[res].GetInt() <= 0 {
				c.flags.z = true
			}

		case opcode.MUL_OP:
			c.ip += 1
			res := c.mem[c.ip]
			c.ip += 1
			a := c.mem[c.ip]
			c.ip += 1
			b := c.mem[c.ip]
			c.ip += 1

			// store result
			aVal := c.regs[a].GetInt()
			bVal := c.regs[b].GetInt()
			c.regs[res].SetInt(aVal * bVal)

		case opcode.DIV_OP:
			c.ip += 1
			res := c.mem[c.ip]
			c.ip += 1
			a := c.mem[c.ip]
			c.ip += 1
			b := c.mem[c.ip]
			c.ip += 1

			// store result
			aVal := c.regs[a].GetInt()
			bVal := c.regs[b].GetInt()

			if bVal == 0 {
				fmt.Printf("Attempting to divide by zero - denying\n")
				os.Exit(3)
			}
			c.regs[res].SetInt(aVal / bVal)

		case opcode.INC_OP:

			// register
			c.ip += 1
			reg := c.mem[c.ip]

			if reg < 0 || reg > 15 {
				fmt.Printf("Register %d out of range\n", reg)
				os.Exit(1)
			}

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

		case opcode.DEC_OP:

			// register
			c.ip += 1
			reg := c.mem[c.ip]

			if reg < 0 || reg > 15 {
				fmt.Printf("Register %d out of range\n", reg)
				os.Exit(1)
			}

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

		case opcode.AND_OP:
			c.ip += 1
			res := c.mem[c.ip]
			c.ip += 1
			a := c.mem[c.ip]
			c.ip += 1
			b := c.mem[c.ip]
			c.ip += 1

			// store result
			aVal := c.regs[a].GetInt()
			bVal := c.regs[b].GetInt()
			c.regs[res].SetInt(aVal & bVal)

		case opcode.OR_OP:
			c.ip += 1
			res := c.mem[c.ip]
			c.ip += 1
			a := c.mem[c.ip]
			c.ip += 1
			b := c.mem[c.ip]
			c.ip += 1

			// store result
			aVal := c.regs[a].GetInt()
			bVal := c.regs[b].GetInt()
			c.regs[res].SetInt(aVal | bVal)

		case opcode.STRING_STORE:
			// register
			c.ip += 1
			reg := c.mem[c.ip]

			if reg < 0 || reg > 15 {
				fmt.Printf("Register %d out of range\n", reg)
				os.Exit(1)
			}

			// bump past that to the length + string
			c.ip += 1

			// read it
			str := c.readString()

			// store the string
			c.regs[reg].SetString(str)

		case opcode.STRING_PRINT:
			// register
			c.ip += 1
			reg := c.mem[c.ip]

			if reg < 0 || reg > 15 {
				fmt.Printf("Register %d out of range\n", reg)
				os.Exit(1)
			}

			fmt.Printf("%s", c.regs[reg].GetString())
			c.ip += 1

		case opcode.STRING_CONCAT:
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

			aVal := c.regs[a].GetString()
			bVal := c.regs[b].GetString()

			c.regs[res].SetString(aVal + bVal)

		case opcode.STRING_SYSTEM:
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

		case opcode.STRING_TOINT:
			// register
			c.ip += 1
			reg := c.mem[c.ip]

			if reg < 0 || reg > 15 {
				fmt.Printf("Register %d out of range\n", reg)
				os.Exit(1)
			}

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

		case opcode.CMP_REG:
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

		case opcode.CMP_IMMEDIATE:
			// register
			c.ip += 1
			reg := int(c.mem[c.ip])

			if reg < 0 || reg > 15 {
				fmt.Printf("Register %d out of range\n", reg)
				os.Exit(1)
			}

			c.ip += 1
			val := c.read2Val()

			if c.regs[reg].Type() == "int" && c.regs[reg].GetInt() == val {
				c.flags.z = true
			} else {
				c.flags.z = false
			}

		case opcode.CMP_STRING:
			// register
			c.ip += 1
			reg := int(c.mem[c.ip])
			if reg < 0 || reg > 15 {
				fmt.Printf("Register %d out of range\n", reg)
				os.Exit(1)
			}

			c.ip += 1

			// read it
			str := c.readString()

			if c.regs[reg].Type() == "string" && c.regs[reg].GetString() == str {
				c.flags.z = true
			} else {
				c.flags.z = false
			}

		case opcode.IS_STRING:
			// register
			c.ip += 1
			reg := int(c.mem[c.ip])
			if reg < 0 || reg > 15 {
				fmt.Printf("Register %d out of range\n", reg)
				os.Exit(1)
			}

			c.ip += 1

			if c.regs[reg].Type() == "string" {
				c.flags.z = true
			} else {
				c.flags.z = false
			}

		case opcode.IS_INTEGER:
			// register
			c.ip += 1
			reg := int(c.mem[c.ip])
			if reg < 0 || reg > 15 {
				fmt.Printf("Register %d out of range\n", reg)
				os.Exit(1)
			}

			c.ip += 1

			if c.regs[reg].Type() == "int" {
				c.flags.z = true
			} else {
				c.flags.z = false
			}

		case opcode.NOP_OP:
			c.ip += 1

		case opcode.REG_STORE:
			// register
			c.ip += 1
			dst := int(c.mem[c.ip])
			c.ip += 1

			// register
			src := int(c.mem[c.ip])
			c.ip += 1

			// Copy the register - paying attention to types
			if c.regs[src].Type() == "string" {
				c.regs[dst].SetString(c.regs[src].GetString())
			} else if c.regs[src].Type() == "int" {
				c.regs[dst].SetInt(c.regs[src].GetInt())
			} else {
				fmt.Printf("Invalid register type?")
				os.Exit(3)
			}

		case opcode.PEEK:
			// register
			c.ip += 1
			result := int(c.mem[c.ip])

			c.ip += 1
			src := int(c.mem[c.ip])

			// get the address from the src register contents
			addr := c.regs[src].GetInt()

			// store the contents of the given address
			c.regs[result].SetInt(int(c.mem[addr]))
			c.ip += 1

		case opcode.POKE:

			// register
			c.ip += 1
			src := int(c.mem[c.ip])
			c.ip += 1

			dst := int(c.mem[c.ip])
			c.ip += 1

			// So the destination will contain an address
			// put the contents of the source to that.
			addr := c.regs[dst].GetInt()
			val := c.regs[src].GetInt()

			c.mem[addr] = byte(val)

		case opcode.MEMCPY:
			// register
			c.ip += 1
			dst := int(c.mem[c.ip])
			c.ip += 1

			src := int(c.mem[c.ip])
			c.ip += 1

			len := int(c.mem[c.ip])
			c.ip += 1

			// get the addresses from the registers
			srcAddr := c.regs[src].GetInt()
			dstAddr := c.regs[dst].GetInt()
			length := c.regs[len].GetInt()

			i := 0
			for i < length {

				if dstAddr >= 0xFFFF {
					dstAddr = 0
				}
				if srcAddr >= 0xFFFF {
					srcAddr = 0
				}

				c.mem[dstAddr] = c.mem[srcAddr]
				dstAddr += 1
				srcAddr += 1
				i += 1
			}

		case opcode.STACK_PUSH:
			// register
			c.ip += 1
			reg := int(c.mem[c.ip])
			if reg < 0 || reg > 15 {
				fmt.Printf("Register %d out of range\n", reg)
				os.Exit(1)
			}

			c.ip += 1

			// Store the value in the register on the stack
			c.stack.Push(c.regs[reg].GetInt())

		case opcode.STACK_POP:
			// register
			c.ip += 1
			reg := int(c.mem[c.ip])
			if reg < 0 || reg > 15 {
				fmt.Printf("Register %d out of range\n", reg)
				os.Exit(1)
			}

			c.ip += 1

			// Ensure our stack isn't empty
			if c.stack.Empty() {
				fmt.Printf("Stack Underflow!\n")
				os.Exit(1)
			}
			// Store the value in the register on the stack
			val, _ := c.stack.Pop()
			c.regs[reg].SetInt(val)

		case opcode.STACK_RET:
			// Ensure our stack isn't empty
			if c.stack.Empty() {
				fmt.Printf("Stack Underflow!\n")
				os.Exit(1)
			}

			// Get the address
			addr, _ := c.stack.Pop()

			// jump
			c.ip = addr

		case opcode.STACK_CALL:
			c.ip += 1

			addr := c.read2Val()

			// push the current IP onto the stack
			c.stack.Push(c.ip)

			// jump to the call address
			c.ip = addr

		case opcode.TRAP_OP:
			c.ip += 1

			num := c.read2Val()

			fn := TRAPS[num]
			if fn != nil {
				fn(c, num)
			}
		default:
			fmt.Printf("Unrecognized/Unimplemented opcode %02X at IP %04X\n", op.Value(), c.ip)
			os.Exit(1)
		}

		// Ensure our instruction-pointer wraps around.
		if c.ip > 0xFFFF {
			c.ip = 0
		}
	}
}
