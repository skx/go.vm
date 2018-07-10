package cpu

import (
	"testing"
)

// Test a new register is empty
func TestNewRegister(t *testing.T) {
	r := NewRegister()
	if r.Type() != "int" {
		t.Errorf("New register is not an int")
	}
	if r.GetInt() != 0 {
		t.Errorf("New register contains a value!")
	}
}

// Test an integer register
func TestRegisterInt(t *testing.T) {
	r := NewRegister()
	r.SetInt(0xffff)

	if r.Type() != "int" {
		t.Errorf("register is not an int")
	}
	if r.GetInt() != 0xffff {
		t.Errorf("register contains the wrong value!")
	}
}

// Test a string register
func TestRegisterString(t *testing.T) {
	r := NewRegister()
	r.SetString("Hello, world!")

	if r.Type() != "string" {
		t.Errorf("register is not a string")
	}
	if r.GetString() != "Hello, world!" {
		t.Errorf("register contains the wrong value!")
	}
}

// Test overflow
func TestIntBounds(t *testing.T) {

	// We expect 0-0xffff
	type TestCase struct {
		set int
		get int
	}

	// Test some negative & excessive values
	tests := []TestCase{
		{-100, 0},
		{-1, 0},
		{0, 0},
		{1, 1},
		{0xffff, 0xffff},
		{0xffffff, 0xffff},
	}

	for _, test := range tests {
		r := NewRegister()
		r.SetInt(test.set)

		if r.Type() != "int" {
			t.Errorf("register is not an int")
		}
		if r.GetInt() != test.get {
			t.Errorf("register contains the wrong value: 0x%04X != 0x%04X", r.GetInt(), test.get)
		}
	}
}
