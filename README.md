[![Go Report Card](https://goreportcard.com/badge/github.com/skx/go.vm)](https://goreportcard.com/report/github.com/skx/go.vm)
[![license](https://img.shields.io/github/license/skx/go.vm.svg)](https://github.com/skx/go.vm/blob/master/LICENSE)
[![Release](https://img.shields.io/github/release/skx/go.vm.svg)](https://github.com/skx/go.vm/releases/latest)

* [go.vm](#govm)
* [Installation](#installation)
  * [Build without Go Modules (Go before 1.11)](#build-without-go-modules-go-before-111)
  * [Build with Go Modules (Go 1.11 or higher)](#build-with-go-modules-go-111-or-higher)
* [Usage](#usage)
* [Opcodes](#opcodes)
* [Notes](#notes)
  * [The compiler](#the-compiler)
  * [The interpreter](#the-interpreter)
  * [Changes](#changes)
  * [DB/DATA Changes](#dbdata-changes)
  * [Traps](#traps)
* [Fuzzing](#fuzzing)
* [See-Also](#see-also)
* [Github Setup](#github-setup)

# go.vm

This project is a golang based compiler and interpreter for a simple virtual
machine.  It is a port of the existing project:

* https://github.com/skx/simple.vm
  * The original project has a perl based compiler/decompiler.
  * The original interpreter was written in C.

You can get a feel for what it looks like by referring to either the parent
project, or [the examples](examples/) contained in this repository.

This particular virtual machine is intentionally simple, but despite that it is hopefully implemented in a readable fashion.  "Simplicity" here means that we support only a small number of instructions, and the 16-registers the virtual CPU possesses can store strings and integers, but not floating-point values.

If you want to see a __real__ virtual machine, interpreting a scripting language, which you can embed inside your Golang applications:

* https://github.com/skx/evalfilter


## Installation

There are two ways to install this project from source, which depend on the version of the [go](https://golang.org/) version you're using.

If you prefer you can fetch a binary from [our release page](https://github.com/skx/go.vm/releases).  Currently there is only a binary for Linux (amd64) due to the use of `cgo` in our dependencies.

## Build without Go Modules (Go before 1.11)

    go get -u github.com/skx/go.vm

## Build with Go Modules (Go 1.11 or higher)

    git clone https://github.com/skx/go.vm ;# make sure to clone outside of GOPATH
    cd go.vm
    go install


## Usage

Once installed there are three sub-commands of interest:

* `go.vm compile $file.in`
   * Compiles the given program into bytecode.
* `go.vm execute $file.raw`
   * Given the path to a file of bytecode, then interpret it.
* `go.vm run $file.in`
   * Compiles the specified program, then directly executes it.

So to compile the input-file `examples/hello.in` into bytecode:

     $ go.vm compile examples/hello.in

Then to execute the resulting bytecode:

     $ go.vm execute examples/hello.raw

Or you can handle both steps at once:

     $ go.vm run examples/hello.in


## Opcodes

The virtual machine has 16 registers, each of which can store an integer
or a string.  For example to set the first two registers you might write:

     store #0, "This is a string"
     store #1, 0xFFFF

In addition to this there are several mathematical operations which have
the general form:

     $operation $result, $src1, $src2

For example to add the contents of register #1 and register #2, storing
the result in register #0 you would write:

     add #0, #1, #2

Strings and integers may be displayed to STDOUT via:

     print_str #1
     print_int #3

Control-flow is supported via `call`, `ret` (for subroutines) and `jmp`
for absolute jumps.  You can also use the `Z`-flag which is set by
comparisons and the `inc`/`dec` instructions and make conditional jumps:

        store #1, 0x42
        cmp #1, 0x42
        jmpz ok

        store #1, "Something weird happened!\n"
        print_str #1
        exit
      :ok
        store #1, "Comparing register #01 to 0x42 succeeded!\n"
        print_str #1
        exit

Further instructions are available and can be viewed beneath [examples/](examples/).  The instruction-set is pretty limited, for example there is no notion of
reading from STDIN - however this _is_ supported via the use of traps, as [documented below](#traps).


## Notes

Some brief notes on parts of the code / operation:

### The compiler

The compiler is built in a traditional fashion:

* Input is split into tokens via [lexer.go](lexer/lexer.go)
  * This uses the [token.go](token/token.go) for the definition of constants.
* The stream of tokens is iterated over by [compiler.go](compiler/compiler.go)
  * This uses the constants in [opcode.go](opcode/opcode.go) for the bytecode generation.

The approach to labels is the same as in the inspiring-project:  Every time
we come across a label we output a pair of temporary bytes in our bytecode.
Later, once we've read the whole program and assume we've found all existing
labels,  we go back up and fix the generated addresses.

You can use the `dump` command to see the structure the lexer generates:

     $ go.vm dump ./examples/hello.in
     {STORE store}
     {IDENT #1}
     {, ,}
     {STRING Hello, World!
     }
     {PRINT_STR print_str}
     {IDENT #1}
     {EXIT exit}


### The interpreter

The core of the interpreter is located in the file [cpu.go](cpu/cpu.go) and is
as simple and naive as you would expect.  There are some supporting files
in the same directory:

* [register.go](cpu/register.go)
  * The implementation of the register-related functions.
* [stack.go](cpu/stack.go)
  * The implementation of the stack.
* [traps.go](cpu/traps.go)
  * The implementation of the traps, to be [described below](#traps).


### Changes

Compared to [the original project](https://github.com/skx/simple.vm) there are two main changes:

* The `DB`/`DATA` operation allows storing string data directly in the generated bytecode.
* There is a notion of `traps`.
   * Rather than defining opcodes for complex tasks it is now possible to callback into the CPU-emulator to do work.

### DB/DATA Changes

For example in simple.vm project this is possible:

    DB 0x01, 0x02,

But this is not:

     DB "This is a string, with terminator to follow"
     DB 0x00

`go.vm` supports this, and it is demonstrated in [examples/peek-strlen.in](examples/peek-strlen.in).

### Traps

The instruction `int` can be used to call back to the emulator to do some work
on behalf of a program.  The following traps are currently defined & available:

* `int 0x00`
   * Set the contents of the register `#0` with the length of the string in register `#0`.
* `int 0x01`
   * Set the contents of the register `#0` with a string entered by the user.
   * See [examples/trap.stdin.in](examples/trap.stdin.in).
* `int 0x02`
   * Update the (string) contents of register `#0` to remove any trailing newline.
   * See [examples/trap.box.in](examples/trap.box.in).

Adding your own trap-functions should be as simple as editing [cpu/traps.go](cpu/traps.go).


## Fuzzing

Fuzz-testing is a powerful technique to discover bugs, in brief it consists
of running a program with numerous random inputs and waiting for it to die.

The CPU in this repository has been fuzzed repeatedly, using the new native fuzz-testing available within go 1.18+.  To run tests:

```sh
cd cpu
$ go test  -parallel=1 -fuzz=FuzzCPU -v
..
=== FUZZ  FuzzCPU
fuzz: elapsed: 0s, gathering baseline coverage: 0/124 completed
fuzz: elapsed: 3s, gathering baseline coverage: 124/124 completed, now fuzzing with 1 workers
fuzz: elapsed: 3s, execs: 124 (41/sec), new interesting: 0 (total: 124)
fuzz: elapsed: 6s, execs: 542 (139/sec), new interesting: 0 (total: 124)
fuzz: elapsed: 9s, execs: 1080 (179/sec), new interesting: 0 (total: 124)
fuzz: elapsed: 12s, execs: 1080 (0/sec), new interesting: 0 (total: 124)
fuzz: elapsed: 15s, execs: 1080 (0/sec), new interesting: 0 (total: 124)
fuzz: elapsed: 18s, execs: 1080 (0/sec), new interesting: 0 (total: 124)
fuzz: elapsed: 21s, execs: 1080 (0/sec), new interesting: 0 (total: 124)
fuzz: elapsed: 24s, execs: 1080 (0/sec), new interesting: 0 (total: 124)
...
```

# See-Also

The original virtual-machine compiler and interpreter is available from the following repository:

* https://github.com/skx/simple.vm

In addition to that you can find a _real_ virtual-machine inside the embedded scripting engine I wrote, also for golang.  In that case a scripting language is parsed and converted into a series of bytecode instructions, which are executed by a virtual machine.  Similar to this project, but the bytecode operations are more complex and high-level:

* https://github.com/skx/evalfilter

If you're interested in compilers, and interpreters, generally you might enjoy these other projects too:

* https://github.com/skx/gobasic
  * A simple BASIC interpreter
* https://github.com/skx/bfcc
  * A [brainfuck](https://en.wikipedia.org/wiki/Brainfuck) compiler
* https://github.com/skx/math-compiler
  * A simple compiler for mathematical expressions.


## Github Setup

This repository is configured to run tests upon every commit, and when
pull-requests are created/updated.  The testing is carried out via
[.github/run-tests.sh](.github/run-tests.sh) which is used by the
[github-action-tester](https://github.com/skx/github-action-tester) action.

Releases are automated in a similar fashion via [.github/build](.github/build),
and the [github-action-publish-binaries](https://github.com/skx/github-action-publish-binaries) action.


Steve
--
