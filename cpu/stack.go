// This file contains the implementation of the stack the CPU uses.
//
// Note that the stack is used for storing integers only.

package cpu

import "errors"

// Stack holds return-addresses when the `call` operation is being
// completed.  It can also be used for storing ints.
type Stack struct {
	// The entries on our stack
	entries []int
}

//
// Stack functions
//

// NewStack creates a new stack object.
func NewStack() *Stack {
	return &Stack{}
}

// Is the stack empty?
func (s *Stack) Empty() bool {
	return (len(s.entries) <= 0)
}

// Get the length of the stack
func (s *Stack) Size() int {
	return (len(s.entries))
}

// Push adds a value to the stack
func (s *Stack) Push(value int) {
	s.entries = append(s.entries, value)
}

// Pop removes a value from the stack
func (s *Stack) Pop() (int, error) {
	if s.Empty() {
		return 0, errors.New("Pop from an empty stack")
	}

	result := s.entries[0]
	s.entries = append(s.entries[:0], s.entries[1:]...)
	return result, nil
}
