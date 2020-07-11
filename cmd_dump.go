package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/google/subcommands"
	"github.com/skx/go.vm/compiler"
	"github.com/skx/go.vm/lexer"
)

type dumpCmd struct {
}

//
// Glue
//
func (*dumpCmd) Name() string     { return "dump" }
func (*dumpCmd) Synopsis() string { return "Show the lexed output of the given program." }
func (*dumpCmd) Usage() string {
	return `dump :
  Demonstrate how our lexer performed by dumping the given input file, as a
  stream of tokens.
`
}

//
// Flag setup: no flags
//
func (p *dumpCmd) SetFlags(f *flag.FlagSet) {
}

//
// Entry-point.
//
func (p *dumpCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	//
	// For each file on the command-line we can dump it.
	//
	for _, file := range f.Args() {

		// Read the file.
		input, err := ioutil.ReadFile(file)
		if err != nil {
			fmt.Printf("Error reading %s - %s\n", file, err.Error())
			return subcommands.ExitFailure
		}

		// Lex it
		l := lexer.New(string(input))

		// Dump it
		e := compiler.New(l)
		e.Dump()
	}
	return subcommands.ExitSuccess
}
