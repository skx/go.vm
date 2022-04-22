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

	val, err := r.GetInt()
	if err != nil {
		t.Errorf("failed to get value")
	}
	if val != 0 {
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
	val, err := r.GetInt()
	if err != nil {
		t.Errorf("failed to get value")
	}
	if val != 0xffff {
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

	val, err := r.GetString()
	if err != nil {
		t.Errorf("error getting string")
	}
	if val != "Hello, world!" {
		t.Errorf("register contains the wrong value!")
	}

	// Calling "GetInt" will fail.
	_, err = r.GetInt()
	if err == nil {
		t.Errorf("expected error, received none")
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

		val, err := r.GetInt()
		if err != nil {
			t.Errorf("failed to get value")
		}

		if val != test.get {
			t.Errorf("register contains the wrong value: 0x%04X != 0x%04X", val, test.get)
		}

		// Calling GetString will fail on an integer value
		_, err = r.GetString()
		if err == nil {
			t.Errorf("expected error, got none")
		}
	}
}
