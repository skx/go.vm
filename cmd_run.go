package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/google/subcommands"
	"github.com/skx/go.vm/compiler"
	"github.com/skx/go.vm/cpu"
	"github.com/skx/go.vm/lexer"
)

type runCmd struct {
}

//
// Glue
//
func (*runCmd) Name() string     { return "run" }
func (*runCmd) Synopsis() string { return "Run the given source program." }
func (*runCmd) Usage() string {
	return `run :
  The run sub-command compiles the given source program, and then executes
  it immediately.
`
}

//
// Flag setup: no flags
//
func (p *runCmd) SetFlags(f *flag.FlagSet) {
}

//
// Entry-point.
//
func (p *runCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	//
	// For each file on the command-line both compile and execute it.
	//
	for _, file := range f.Args() {
		fmt.Printf("Parsing file: %s\n", file)

		// Read the file.
		input, err := ioutil.ReadFile(file)
		if err != nil {
			fmt.Printf("Error reading %s - %s\n", file, err.Error())
			return subcommands.ExitFailure
		}

		// Lex it
		l := lexer.New(string(input))

		// Compile it.
		e := compiler.New(l)
		e.Compile()

		// Now create a machine to run the compiled program in
		c := cpu.NewCPU()

		// Load the program
		c.LoadBytes(e.Output())

		// Run the machine
		err = c.Run()
		if err != nil {
			fmt.Printf("Error running file: %s\n", err)
			return subcommands.ExitFailure
		}
	}
	return subcommands.ExitSuccess
}
