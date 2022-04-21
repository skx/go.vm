package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/google/subcommands"
	"github.com/skx/go.vm/cpu"
)

type executeCmd struct {
}

//
// Glue
//
func (*executeCmd) Name() string     { return "execute" }
func (*executeCmd) Synopsis() string { return "Executed a compiled program." }
func (*executeCmd) Usage() string {
	return `execute :
  Execute the bytecodes contained in the given input file.
`
}

//
// Flag setup: no flags
//
func (p *executeCmd) SetFlags(f *flag.FlagSet) {
}

//
// Entry-point.
//
func (p *executeCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	//
	// For each file on the command-line we can now execute it.
	//
	for _, file := range f.Args() {

		c := cpu.NewCPU()

		err := c.LoadFile(file)
		if err != nil {
			fmt.Printf("Error loading file: %s\n", err)
		}

		err = c.Run()
		if err != nil {
			fmt.Printf("Error running file: %s\n", err)
			return subcommands.ExitFailure
		}
	}
	return subcommands.ExitSuccess
}
