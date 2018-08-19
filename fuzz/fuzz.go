// +build gofuzz

package fuzz

import (
	"github.com/skx/go.vm/cpu"
)

func Fuzz(data []byte) int {
	c := cpu.NewCPU()
	c.LoadBytes(data)
	c.Run()
	return 0
}
