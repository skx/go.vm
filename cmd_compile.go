package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/google/subcommands"
	"github.com/skx/go.vm/compiler"
	"github.com/skx/go.vm/lexer"
)

type compileCmd struct {
}

//
// Glue
//
func (*compileCmd) Name() string     { return "compile" }
func (*compileCmd) Synopsis() string { return "Compile a simple.vm program." }
func (*compileCmd) Usage() string {
	return `compile :
  Compile the given input file to a series of bytecodes.
`
}

//
// Flag setup: no flags
//
func (p *compileCmd) SetFlags(f *flag.FlagSet) {
}

//
// Entry-point.
//
func (p *compileCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	//
	// For each file on the command-line we can compile it.
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

		// Compile it
		e := compiler.New(l)
		e.Compile()

		// Write it out - remove the suffix from the file
		name := strings.TrimSuffix(file, filepath.Ext(file))

		// Add a .raw suffix to the file.
		e.Write(name + ".raw")
	}
	return subcommands.ExitSuccess
}
