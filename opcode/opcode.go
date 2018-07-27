package opcode

var (
	// EXIT is our first opcode.
	// The rest follow, grouped by functionality.
	EXIT          = 0x00
	INT_STORE     = 0x01
	INT_PRINT     = 0x02
	INT_TOSTRING  = 0x03
	INT_RANDOM    = 0x04
	JUMP_TO       = 0x10
	JUMP_Z        = 0x11
	JUMP_NZ       = 0x12
	XOR_OP        = 0x20
	ADD_OP        = 0x21
	SUB_OP        = 0x22
	MUL_OP        = 0x23
	DIV_OP        = 0x24
	INC_OP        = 0x25
	DEC_OP        = 0x26
	AND_OP        = 0x27
	OR_OP         = 0x28
	STRING_STORE  = 0x30
	STRING_PRINT  = 0x31
	STRING_CONCAT = 0x32
	STRING_SYSTEM = 0x33
	STRING_TOINT  = 0x34
	CMP_REG       = 0x40
	CMP_IMMEDIATE = 0x41
	CMP_STRING    = 0x42
	IS_STRING     = 0x43
	IS_INTEGER    = 0x44
	NOP_OP        = 0x50
	REG_STORE     = 0x51
	PEEK          = 0x60
	POKE          = 0x61
	MEMCPY        = 0x62
	STACK_PUSH    = 0x70
	STACK_POP     = 0x71
	STACK_RET     = 0x72
	STACK_CALL    = 0x73
	TRAP_OP       = 0x80
)

// Opcode is a holder for a single instruction.
// Note that this doesn't take any account of the arguments which might
// be necessary.
type Opcode struct {
	instruction byte
}

// NewOpcode creates a new Opcode.
func NewOpcode(instruction byte) *Opcode {
	o := &Opcode{}
	o.instruction = instruction
	return o
}

// String converts the given Opcode to a string, but again note that it
// doesn't take into account the value.
func (o *Opcode) String() string {
	switch int(o.instruction) {
	case EXIT:
		return "exit"
	case INT_STORE:
		return "INT_STORE"
	case INT_PRINT:
		return "INT_PRINT"
	case INT_TOSTRING:
		return "INT_TOSTRING"
	case INT_RANDOM:
		return "INT_RANDOM"
	case JUMP_TO:
		return "JUMP_TO"
	case JUMP_Z:
		return "JUMP_Z"
	case JUMP_NZ:
		return "JUMP_NZ"

	case XOR_OP:
		return "XOR_OP"
	case ADD_OP:
		return "ADD_OP"
	case SUB_OP:
		return "SUB_OP"
	case MUL_OP:
		return "MUL_OP"
	case DIV_OP:
		return "DIV_OP"
	case INC_OP:
		return "INC_OP"
	case DEC_OP:
		return "DEC_OP"
	case AND_OP:
		return "AND_OP"
	case OR_OP:
		return "OR_OP"
	case STRING_STORE:
		return "STRING_STORE"
	case STRING_PRINT:
		return "STRING_PRINT"
	case STRING_CONCAT:
		return "STRING_CONCAT"
	case STRING_SYSTEM:
		return "STRING_SYSTEM"
	case STRING_TOINT:
		return "STRING_TOINT"
	case CMP_REG:
		return "CMP_REG"
	case CMP_IMMEDIATE:
		return "CMP_IMMEDIATE"
	case CMP_STRING:
		return "CMP_STRING"
	case IS_STRING:
		return "IS_STRING"
	case IS_INTEGER:
		return "IS_INTEGER"
	case NOP_OP:
		return "NOP"
	case REG_STORE:
		return "REG_STORE"
	case PEEK:
		return "PEEK"
	case POKE:
		return "POKE"
	case MEMCPY:
		return "MEMCPY"
	case STACK_PUSH:
		return "PUSH"
	case STACK_POP:
		return "POP"
	case STACK_RET:
		return "RET"
	case STACK_CALL:
		return "CALL"
	case TRAP_OP:
		return "TRAP"
	}
	return "UNKNOWN OPCODE .."
}

// Value returns the byte-value of the opcode.
func (o *Opcode) Value() byte {
	return (o.instruction)
}
