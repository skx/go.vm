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
	r.SetInt(17)

	if r.Type() != "int" {
		t.Errorf("register is not an int")
	}
	if r.GetInt() != 17 {
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
