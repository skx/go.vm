package opcode

var (
	// Exit
	EXIT = 0x00

	// Int operations

	INT_STORE    = 0x01
	INT_PRINT    = 0x02
	INT_TOSTRING = 0x03
	INT_RANDOM   = 0x04

	//  Jumps
	JUMP_TO = 0x10
	JUMP_Z  = 0x11
	JUMP_NZ = 0x12

	//  Mathematical

	XOR_OP = 0x20
	ADD_OP = 0x21
	SUB_OP = 0x22
	MUL_OP = 0x23
	DIV_OP = 0x24
	INC_OP = 0x25
	DEC_OP = 0x26
	AND_OP = 0x27
	OR_OP  = 0x28

	//  String operations
	STRING_STORE  = 0x30
	STRING_PRINT  = 0x31
	STRING_CONCAT = 0x32
	STRING_SYSTEM = 0x33
	STRING_TOINT  = 0x34

	//  Comparison functions

	CMP_REG       = 0x40
	CMP_IMMEDIATE = 0x41
	CMP_STRING    = 0x42
	IS_STRING     = 0x43
	IS_INTEGER    = 0x44

	//  Misc things

	NOP_OP    = 0x50
	REG_STORE = 0x51

	//  Load from RAM/store in RAM

	PEEK   = 0x60
	POKE   = 0x61
	MEMCPY = 0x62

	//  Stack operations

	STACK_PUSH = 0x70
	STACK_POP  = 0x71
	STACK_RET  = 0x72
	STACK_CALL = 0x73

	// Interrupt / trap
	TRAP_OP = 0x80
)
