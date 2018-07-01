package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/google/subcommands"
)

//
// modified during testing
//
var out io.Writer = os.Stdout

var (
	version = "master"
)

type versionCmd struct {
}

//
// Glue
//
func (*versionCmd) Name() string     { return "version" }
func (*versionCmd) Synopsis() string { return "Show our version." }
func (*versionCmd) Usage() string {
	return `version :
  Report upon our version, and exit.
`
}

//
// Flag setup.
//
func (p *versionCmd) SetFlags(f *flag.FlagSet) {
}

//
// Show the version - using the "out"-writer.
//
func showVersion() {
	fmt.Fprintf(out, "%s\n", version)
}

//
// Entry-point.
//
func (p *versionCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	showVersion()
	return subcommands.ExitSuccess
}
