package cpu

import (
	"testing"
)

// Test a new stack is empty
func TestStackStartsEmpty(t *testing.T) {
	s := NewStack()
	if !s.Empty() {
		t.Errorf("New stack is non-empty")
	}
	if s.Size() != 0 {
		t.Errorf("New stack is non-empty")
	}
}

// Test we can add/remove a value
func TestStack(t *testing.T) {
	s := NewStack()

	s.Push(42)

	if s.Empty() {
		t.Errorf("Stack should not be empty after adding item.")
	}
	if s.Size() != 1 {
		t.Errorf("stack has a size-mismatch")
	}

	val, err := s.Pop()

	if err != nil {
		t.Errorf("Received an unexpected error popping from the stack")
	}
	if !s.Empty() {
		t.Errorf("Stack should be empty now.")
	}
	if s.Size() != 0 {
		t.Errorf("stack has a size-mismatch")
	}

	if val != 42 {
		t.Errorf("Stack push/pop mismatch")
	}
}

// Popping from an empty stack should fail
func TestEmptyStack(t *testing.T) {
	s := NewStack()

	_, err := s.Pop()

	if err == nil {
		t.Errorf("should receive an error popping an empty stack!")
	}
}


// Test issue #12 - stack is FIFO, not LIFO
func TestIssue12(t *testing.T) {

	s := NewStack()
	s.Push(10)  // top is 10
	s.Push(20)  // top is 20, then 10
	s.Push(30)  // top is 30, then 20, then 10

	// Ensure the contents are as expected
	if s.entries[0] != 10 { t.Fatalf("Unexpected result")}
	if s.entries[1] != 20 { t.Fatalf("Unexpected result")}
	if s.entries[2] != 30 { t.Fatalf("Unexpected result")}
	if s.Size() != 3 { t.Fatalf("wrong length") }


	// popping should remove in expected order
	val,err := s.Pop()
	if err != nil {
		t.Fatalf("unexpected error")
	}
	if val != 30 {
		t.Fatalf("stack is wrong")
	}

	// Contents should still be what we expect,
	// after removing one entry
	if s.entries[0] != 10 { t.Fatalf("Unexpected result")}
	if s.entries[1] != 20 { t.Fatalf("Unexpected result")}
	if s.Size() != 2 { t.Fatalf("wrong length") }

	// Get the middle value
	val,err = s.Pop()
	if err != nil {
		t.Fatalf("unexpected error")
	}
	if val != 20 {
		t.Fatalf("stack is wrong")
	}


	if s.entries[0] != 10 { t.Fatalf("Unexpected result")}
	if s.Size() != 1 { t.Fatalf("wrong length")}
	val,err = s.Pop()
	if err != nil {
		t.Fatalf("unexpected error")
	}
	if val != 10 {
		t.Fatalf("stack is wrong")
	}

	if !s.Empty() {
		t.Fatalf("stack should be empty")
	}
	if s.Size() != 0 { t.Fatalf("wrong length")}
}
