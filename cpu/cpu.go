// Package cpu contains the CPU for our virtual machine interpreter.
//
// We should probably use the constants defined in `opcodes/opcodes.go`
// instead of the literal hex-constants for our bytecode, but that's a minor
// issue.
//
package cpu

import (
	"bufio"
	"bytes"
	"context"
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
	regs [15]*Register

	// Flags
	flags Flags

	// Our RAM - where the program is loaded
	mem [0xFFFF]byte

	// Instruction-pointer
	ip int

	// stack
	stack *Stack

	// context is used by callers to implement timeouts.
	context context.Context

	// STDIN is an input-reader used for the input-trap.
	STDIN *bufio.Reader

	// STDOUT is the writer used for outputing things.
	STDOUT *bufio.Writer
}

//
// CPU / VM functions
//

// NewCPU returns a new CPU object.
func NewCPU() *CPU {
	x := &CPU{context: context.Background()}
	x.Reset()

	// allow reading from STDIN
	x.STDIN = bufio.NewReader(os.Stdin)

	// set standard output for STDOUT
	x.STDOUT = bufio.NewWriter(os.Stdout)

	return x
}

// SetContext allows a context to be used as our virtual machine is
// running. This is most used to allow our caller to setup a
// timeout/deadline which will avoid denial-of-service problems if
// user-supplied script(s) contain infinite loops.
func (c *CPU) SetContext(ctx context.Context) {
	c.context = ctx
}

// Reset sets the CPU into a known-good state, by setting the IP to zero,
// and emptying all registers (i.e. setting them to zero too).
func (c *CPU) Reset() {

	// Reset registers
	for i := 0; i < len(c.regs); i++ {
		c.regs[i] = NewRegister()
	}

	// Reset stack
	c.stack = NewStack()

	// Reset instruction pointer to zero.
	c.ip = 0
}

// LoadFile loads the program from the named file into RAM.
// NOTE: The CPU-state is reset prior to the load.
func (c *CPU) LoadFile(path string) error {

	// Load the file
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %s - %s", path, err.Error())
	}

	if len(b) >= 0xFFFF {
		return fmt.Errorf("program too large for RAM %d", len(b))
	}

	// Copy contents of file to our memory region.
	// NOTE: This calls `Reset` too :)
	c.LoadBytes(b)
	return nil
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
func (c *CPU) readString() (string, error) {

	// Read the length of the string we expect
	len := c.read2Val()

	// Can't read beyond RAM, but we will allow wrap-around.
	if len >= 0xffff {
		return "", fmt.Errorf("string too large")
	}

	addr := c.ip

	// Now build up the body of the string
	s := ""
	for i := 0; i < len; i++ {

		tmp := addr + i

		// wrap around
		if tmp == 0xFFFF {
			tmp = 0
		}
		s += string(c.mem[tmp])
	}

	// Jump the IP over the length of the string.
	c.ip += (len)
	return s, nil
}

// Read a two-byte number from the current IP.
// i.e This reads two bytes and returns a 16-bit value to the caller,
// skipping over both bytes in the IP.
func (c *CPU) read2Val() int {
	l := int(c.mem[c.ip])
	c.ip++
	h := int(c.mem[c.ip])
	c.ip++

	val := l + h*256
	return (val)
}

// Run launches our intepreter.
// It does not terminate until an `EXIT` instruction is hit.
func (c *CPU) Run() error {
	run := true
	for run {

		if c.ip >= 0xffff {
			return fmt.Errorf("reading beyond RAM")
		}

		op := opcode.NewOpcode(c.mem[c.ip])
		debugPrintf("%04X %02X [%s]\n", c.ip, op.Value(), op.String())

		//
		// We've been given a context, which we'll test at every
		// iteration of our main-loop.
		//
		// This is a little slow and inefficient, but we need
		// to allow our execution to be time-limited.
		//
		select {
		case <-c.context.Done():
			return fmt.Errorf("timeout during execution")
		default:
			// nop
		}

		switch int(op.Value()) {
		case opcode.EXIT:
			run = false

		case opcode.INT_STORE:
			// register
			c.ip++
			reg := int(c.mem[c.ip])

			// bounds-check our register
			if reg >= len(c.regs) {
				return fmt.Errorf("register %d out of range", reg)
			}

			c.ip++
			val := c.read2Val()
			c.regs[reg].SetInt(val)

		case opcode.INT_PRINT:
			// register
			c.ip++
			reg := c.mem[c.ip]

			// bounds-check our register
			if int(reg) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", reg)
			}

			val, err := c.regs[reg].GetInt()
			if err != nil {
				return err
			}
			if val < 256 {
				c.STDOUT.WriteString(fmt.Sprintf("%02X", val))
			} else {
				c.STDOUT.WriteString(fmt.Sprintf("%04X", val))
			}
			c.STDOUT.Flush()
			c.ip++

		case opcode.INT_TOSTRING:
			// register
			c.ip++
			reg := c.mem[c.ip]

			// bounds-check our register
			if int(reg) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", reg)
			}

			// get value
			i, err := c.regs[reg].GetInt()
			if err != nil {
				return err
			}

			// change from int to string
			c.regs[reg].SetString(fmt.Sprintf("%d", i))

			// next instruction
			c.ip++

		case opcode.INT_RANDOM:
			// register
			c.ip++
			reg := c.mem[c.ip]

			// bounds-check our register
			if int(reg) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", reg)
			}

			// New random source
			s1 := rand.NewSource(time.Now().UnixNano())
			r1 := rand.New(s1)

			// New random number
			c.regs[reg].SetInt(r1.Intn(0xffff))
			c.ip++

		case opcode.JUMP_TO:
			c.ip++
			addr := c.read2Val()
			c.ip = addr

		case opcode.JUMP_Z:
			c.ip++
			addr := c.read2Val()
			if c.flags.z {
				c.ip = addr
			}

		case opcode.JUMP_NZ:
			c.ip++
			addr := c.read2Val()
			if !c.flags.z {
				c.ip = addr
			}

		case opcode.XOR_OP:
			c.ip++
			res := c.mem[c.ip]
			c.ip++
			a := c.mem[c.ip]
			c.ip++
			b := c.mem[c.ip]
			c.ip++

			if int(a) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", a)
			}
			if int(b) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", b)
			}
			if int(res) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", res)
			}

			// store result
			aVal, aErr := c.regs[a].GetInt()
			if aErr != nil {
				return aErr
			}
			bVal, bErr := c.regs[b].GetInt()
			if bErr != nil {
				return bErr
			}
			c.regs[res].SetInt(aVal ^ bVal)

		case opcode.ADD_OP:
			c.ip++
			res := c.mem[c.ip]
			c.ip++
			a := c.mem[c.ip]
			c.ip++
			b := c.mem[c.ip]
			c.ip++

			if int(a) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", a)
			}
			if int(b) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", b)
			}
			if int(res) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", res)
			}

			// store result
			aVal, aErr := c.regs[a].GetInt()
			if aErr != nil {
				return aErr
			}
			bVal, bErr := c.regs[b].GetInt()
			if bErr != nil {
				return bErr
			}
			c.regs[res].SetInt(aVal + bVal)

		case opcode.SUB_OP:
			c.ip++
			res := c.mem[c.ip]
			c.ip++
			a := c.mem[c.ip]
			c.ip++
			b := c.mem[c.ip]
			c.ip++

			if int(a) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", a)
			}
			if int(b) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", b)
			}
			if int(res) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", res)
			}

			// store result
			aVal, aErr := c.regs[a].GetInt()
			if aErr != nil {
				return aErr
			}
			bVal, bErr := c.regs[b].GetInt()
			if bErr != nil {
				return bErr
			}
			c.regs[res].SetInt(aVal - bVal)

			// set the zero-flag if the result was zero or less
			rVal, err := c.regs[res].GetInt()
			if err != nil {
				return err
			}
			if rVal <= 0 {
				c.flags.z = true
			}

		case opcode.MUL_OP:
			c.ip++
			res := c.mem[c.ip]
			c.ip++
			a := c.mem[c.ip]
			c.ip++
			b := c.mem[c.ip]
			c.ip++

			if int(a) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", a)
			}
			if int(b) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", b)
			}
			if int(res) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", res)
			}

			// store result
			aVal, aErr := c.regs[a].GetInt()
			if aErr != nil {
				return aErr
			}
			bVal, bErr := c.regs[b].GetInt()
			if bErr != nil {
				return bErr
			}
			c.regs[res].SetInt(aVal * bVal)

		case opcode.DIV_OP:
			c.ip++
			res := c.mem[c.ip]
			c.ip++
			a := c.mem[c.ip]
			c.ip++
			b := c.mem[c.ip]
			c.ip++

			if int(a) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", a)
			}
			if int(b) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", b)
			}
			if int(res) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", res)
			}

			// store result
			aVal, aErr := c.regs[a].GetInt()
			if aErr != nil {
				return aErr
			}
			bVal, bErr := c.regs[b].GetInt()
			if bErr != nil {
				return bErr
			}

			if bVal == 0 {
				return fmt.Errorf("attempted division by zero")
			}
			c.regs[res].SetInt(aVal / bVal)

		case opcode.INC_OP:

			// register
			c.ip++
			reg := c.mem[c.ip]

			// bounds-check our register
			if int(reg) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", reg)
			}

			// get the value
			val, err := c.regs[reg].GetInt()
			if err != nil {
				return err
			}

			// if the value is the max it will wrap around
			if val == 0xFFFF {
				val = 0
			} else {
				// otherwise be incremented normally
				val++
			}

			// zero?
			c.flags.z = (val == 0)

			c.regs[reg].SetInt(val)

			// bump past that
			c.ip++

		case opcode.DEC_OP:

			// register
			c.ip++
			reg := c.mem[c.ip]

			// bounds-check our register
			if int(reg) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", reg)
			}

			// get the value
			val, err := c.regs[reg].GetInt()
			if err != nil {
				return err
			}

			// if the value is the minimum it will wrap around
			if val == 0x0000 {
				val = 0xFFFF
			} else {
				// otherwise decrease normally
				val--
			}

			// zero?
			c.flags.z = (val == 0)

			c.regs[reg].SetInt(val)

			// bump past that
			c.ip++

		case opcode.AND_OP:
			c.ip++
			res := c.mem[c.ip]
			c.ip++
			a := c.mem[c.ip]
			c.ip++
			b := c.mem[c.ip]
			c.ip++

			if int(a) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", a)
			}
			if int(b) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", b)
			}
			if int(res) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", res)
			}

			// store result
			aVal, aErr := c.regs[a].GetInt()
			if aErr != nil {
				return aErr
			}
			bVal, bErr := c.regs[b].GetInt()
			if bErr != nil {
				return bErr
			}
			c.regs[res].SetInt(aVal & bVal)

		case opcode.OR_OP:
			c.ip++
			res := c.mem[c.ip]
			c.ip++
			a := c.mem[c.ip]
			c.ip++
			b := c.mem[c.ip]
			c.ip++

			if int(a) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", a)
			}
			if int(b) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", b)
			}
			if int(res) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", res)
			}

			// store result
			aVal, aErr := c.regs[a].GetInt()
			if aErr != nil {
				return aErr
			}
			bVal, bErr := c.regs[b].GetInt()
			if bErr != nil {
				return bErr
			}
			c.regs[res].SetInt(aVal | bVal)

		case opcode.STRING_STORE:
			// register
			c.ip++
			reg := c.mem[c.ip]

			// bounds-check our register
			if int(reg) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", reg)
			}

			// bump past that to the length + string
			c.ip++

			// read it
			str, err := c.readString()
			if err != nil {
				return err
			}

			// store the string
			c.regs[reg].SetString(str)

		case opcode.STRING_PRINT:
			// register
			c.ip++
			reg := c.mem[c.ip]

			// bounds-check our register
			if int(reg) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", reg)
			}

			str, err := c.regs[reg].GetString()
			if err != nil {
				return err
			}
			c.STDOUT.WriteString(str)
			c.STDOUT.Flush()
			c.ip++

		case opcode.STRING_CONCAT:
			// output register
			c.ip++
			res := c.mem[c.ip]

			// src1
			c.ip++
			a := c.mem[c.ip]

			if int(a) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", a)
			}

			// src2
			c.ip++
			b := c.mem[c.ip]
			if int(b) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", b)
			}

			c.ip++

			if int(a) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", a)
			}
			if int(b) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", b)
			}
			if int(res) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", res)
			}

			aVal, aErr := c.regs[a].GetString()
			if aErr != nil {
				return aErr
			}
			bVal, bErr := c.regs[b].GetString()
			if bErr != nil {
				return bErr
			}

			c.regs[res].SetString(aVal + bVal)

		case opcode.STRING_SYSTEM:
			// register
			c.ip++
			r := c.mem[c.ip]
			c.ip++

			if int(r) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", r)
			}

			str, sErr := c.regs[r].GetString()
			if sErr != nil {
				return sErr
			}

			if false {
				// run the command
				toExec := splitCommand(str)
				cmd := exec.Command(toExec[0], toExec[1:]...)

				var out bytes.Buffer
				var err bytes.Buffer
				cmd.Stdout = &out
				cmd.Stderr = &err
				er := cmd.Run()
				if er != nil {
					return fmt.Errorf("error invoking system(%s): %s", str, er)
				}

				// stdout
				fmt.Printf("%s", out.String())

				// stderr - if non-empty
				if len(err.String()) > 0 {
					fmt.Printf("%s", err.String())
				}
			}
		case opcode.STRING_TOINT:
			// register
			c.ip++
			reg := c.mem[c.ip]

			// bounds-check our register
			if int(reg) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", reg)
			}

			// get value
			s, sErr := c.regs[reg].GetString()
			if sErr != nil {
				return sErr
			}

			i, err := strconv.Atoi(s)
			if err == nil {
				c.regs[reg].SetInt(i)
			} else {
				return fmt.Errorf("failed to convert %s to int:%s", s, err)
			}

			// next instruction
			c.ip++

		case opcode.CMP_REG:
			c.ip++
			r1 := int(c.mem[c.ip])
			c.ip++
			r2 := int(c.mem[c.ip])
			c.ip++

			if int(r1) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", r1)
			}
			if int(r2) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", r2)
			}

			c.flags.z = false

			switch c.regs[r1].Type() {
			case "int":

				aVal, aErr := c.regs[r1].GetInt()
				if aErr != nil {
					return aErr
				}
				bVal, bErr := c.regs[r2].GetInt()
				if bErr != nil {
					return bErr
				}

				if aVal == bVal {
					c.flags.z = true
				}
			case "string":

				aVal, aErr := c.regs[r1].GetString()
				if aErr != nil {
					return aErr
				}
				bVal, bErr := c.regs[r2].GetString()
				if bErr != nil {
					return bErr
				}

				if aVal == bVal {
					c.flags.z = true
				}
			}

		case opcode.CMP_IMMEDIATE:
			// register
			c.ip++
			reg := int(c.mem[c.ip])

			// bounds-check our register
			if int(reg) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", reg)
			}

			c.ip++
			val := c.read2Val()

			if c.regs[reg].Type() == "int" {
				valCur, err := c.regs[reg].GetInt()
				if err != nil {
					return err
				}
				if valCur == val {
					c.flags.z = true
				}
			} else {
				c.flags.z = false
			}

		case opcode.CMP_STRING:
			// register
			c.ip++
			reg := int(c.mem[c.ip])

			// bounds-check our register
			if int(reg) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", reg)
			}

			c.ip++

			// read it
			str, err := c.readString()
			if err != nil {
				return err
			}

			if c.regs[reg].Type() == "string" {
				val, err := c.regs[reg].GetString()
				if err != nil {
					return err
				}
				if val == str {
					c.flags.z = true
				}
			} else {
				c.flags.z = false
			}

		case opcode.IS_STRING:
			// register
			c.ip++
			reg := int(c.mem[c.ip])

			// bounds-check our register
			if int(reg) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", reg)
			}

			c.ip++

			if c.regs[reg].Type() == "string" {
				c.flags.z = true
			} else {
				c.flags.z = false
			}

		case opcode.IS_INTEGER:
			// register
			c.ip++
			reg := int(c.mem[c.ip])

			// bounds-check our register
			if int(reg) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", reg)
			}

			c.ip++

			if c.regs[reg].Type() == "int" {
				c.flags.z = true
			} else {
				c.flags.z = false
			}

		case opcode.NOP_OP:
			c.ip++

		case opcode.REG_STORE:
			// register
			c.ip++
			dst := int(c.mem[c.ip])
			c.ip++

			// register
			src := int(c.mem[c.ip])
			c.ip++

			if int(src) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", src)
			}
			if int(dst) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", dst)
			}

			// Copy the register - paying attention to types
			if c.regs[src].Type() == "string" {
				cur, err := c.regs[src].GetString()
				if err != nil {
					return err
				}

				c.regs[dst].SetString(cur)
			} else if c.regs[src].Type() == "int" {
				cur, err := c.regs[src].GetInt()
				if err != nil {
					return err
				}
				c.regs[dst].SetInt(cur)
			} else {
				return fmt.Errorf("invalid register type?")
			}

		case opcode.PEEK:
			// register
			c.ip++
			result := int(c.mem[c.ip])

			c.ip++
			src := int(c.mem[c.ip])

			if int(src) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", src)
			}
			if int(result) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", result)
			}

			// get the address from the src register contents
			addr, err := c.regs[src].GetInt()
			if err != nil {
				return err
			}

			if addr >= 0xFFFF {
				return fmt.Errorf("address out of range %d", addr)
			}

			// store the contents of the given address
			c.regs[result].SetInt(int(c.mem[addr]))
			c.ip++

		case opcode.POKE:

			// register
			c.ip++
			src := int(c.mem[c.ip])
			c.ip++

			dst := int(c.mem[c.ip])
			c.ip++

			if int(src) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", src)
			}
			if int(dst) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", dst)
			}

			// So the destination will contain an address
			// put the contents of the source to that.
			addr, err := c.regs[dst].GetInt()
			if err != nil {
				return err
			}

			if addr >= 0xFFFF {
				return fmt.Errorf("address out of range %d", addr)
			}

			val, err2 := c.regs[src].GetInt()
			if err2 != nil {
				return err2
			}

			if addr >= 0xffff {
				return fmt.Errorf("attempting to write beyond RAM")
			}
			c.mem[addr] = byte(val)

		case opcode.MEMCPY:
			// register
			c.ip++
			dst := int(c.mem[c.ip])
			c.ip++

			src := int(c.mem[c.ip])
			c.ip++

			ln := int(c.mem[c.ip])
			c.ip++

			if int(src) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", src)
			}
			if int(dst) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", dst)
			}
			if int(ln) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", ln)
			}

			// get the addresses from the registers
			srcAddr, sErr := c.regs[src].GetInt()
			if sErr != nil {
				return sErr
			}
			dstAddr, dErr := c.regs[dst].GetInt()
			if dErr != nil {
				return dErr
			}
			length, lErr := c.regs[ln].GetInt()
			if lErr != nil {
				return lErr
			}

			i := 0
			for i < length {

				if dstAddr >= 0xFFFF {
					dstAddr = 0
				}
				if srcAddr >= 0xFFFF {
					srcAddr = 0
				}

				c.mem[dstAddr] = c.mem[srcAddr]
				dstAddr++
				srcAddr++
				i++
			}

		case opcode.STACK_PUSH:
			// register
			c.ip++
			reg := int(c.mem[c.ip])

			// bounds-check our register
			if int(reg) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", reg)
			}

			c.ip++

			// Store the value in the register on the stack
			cur, err := c.regs[reg].GetInt()
			if err != nil {
				return err
			}
			c.stack.Push(cur)

		case opcode.STACK_POP:
			// register
			c.ip++
			reg := int(c.mem[c.ip])

			// bounds-check our register
			if int(reg) >= len(c.regs) {
				return fmt.Errorf("register %d out of range", reg)
			}

			c.ip++

			// Ensure our stack isn't empty
			if c.stack.Empty() {
				return fmt.Errorf("stackunderflow")
			}
			// Store the value in the register on the stack
			val, _ := c.stack.Pop()
			c.regs[reg].SetInt(val)

		case opcode.STACK_RET:
			// Ensure our stack isn't empty
			if c.stack.Empty() {
				return fmt.Errorf("stackunderflow")
			}

			// Get the address
			addr, _ := c.stack.Pop()

			// jump
			c.ip = addr

		case opcode.STACK_CALL:
			c.ip++

			addr := c.read2Val()

			// push the current IP onto the stack
			c.stack.Push(c.ip)

			// jump to the call address
			c.ip = addr

		case opcode.TRAP_OP:
			c.ip++

			num := c.read2Val()

			if num < 0 || num >= 0xffff {
				return fmt.Errorf("invalid trap number %d", num)
			}

			fn := TRAPS[num]
			if fn != nil {
				err := fn(c, num)
				if err != nil {
					return err
				}
			}
		default:
			return fmt.Errorf("unrecognized/Unimplemented opcode %02X at IP %04X", op.Value(), c.ip)
		}

		// Ensure our instruction-pointer wraps around.
		if c.ip > 0xFFFF {
			c.ip = 0
		}
	}

	return nil
}
