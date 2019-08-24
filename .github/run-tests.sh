#!/bin/sh

# Install the lint-tool, and the shadow-tool
go get -u golang.org/x/lint/golint
go get -u golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow
go get -u honnef.co/go/tools/cmd/staticcheck


# Run the static-check tool - we ignore errors relating to case
t=$(mktemp)
staticcheck -checks all ./... | grep -v ALL_CAPS > $t
if [ -s $t ]; then
    echo "Found errors via 'staticcheck':"
    cat $t
    rm $t
    exit 1
fi
rm $t


# At this point failures cause aborts
set -e

# Run the linter
echo "Launching linter .."

#
# We have a bunch of errors which we need to mask
#
#  opcode/opcode.go:8:2: don't use ALL_CAPS in Go names; use CamelCase
#  opcode/opcode.go:11:2: don't use ALL_CAPS in Go names; use CamelCase
#  opcode/opcode.go:14:2: don't use ALL_CAPS in Go names; use CamelCase
#  opcode/opcode.go:17:2: don't use ALL_CAPS in Go names; use CamelCase
#  opcode/opcode.go:20:2: don't use ALL_CAPS in Go names; use CamelCase
#
( golint  ./...  | grep -v ALL_CAPS > lint.out ) || true
if [ -s lint.out ]; then
    echo "Linter errors: "
    cat lint.out
    exit 0
else
    rm lint.out
fi

echo "Completed linter .."

# Run the shadow-checker
echo "Launching shadowed-variable check .."
go vet -vettool=$(which shadow) ./...
echo "Completed shadowed-variable check .."

# Run golang tests
go test ./...
