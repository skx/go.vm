package lexer

import (
	"testing"

	"github.com/skx/go.vm/token"
)

func TestNextTokenTrivial(t *testing.T) {
	input := `,`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.COMMA, ","},
		{token.EOF, ""},
	}
	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong, expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - Literal wrong, expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextTokenReal(t *testing.T) {
	input := `
        store #1, 0x0a
        store #2, 0xFF
        add #0, #1, #2
        print_int #0
        store #1, "steve"
        print_str #1
`
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.STORE, "store"},
		{token.IDENT, "#1"},
		{token.COMMA, ","},
		{token.INT, "0x0a"},

		{token.STORE, "store"},
		{token.IDENT, "#2"},
		{token.COMMA, ","},
		{token.INT, "0xFF"},

		{token.ADD, "add"},
		{token.IDENT, "#0"},
		{token.COMMA, ","},
		{token.IDENT, "#1"},
		{token.COMMA, ","},
		{token.IDENT, "#2"},

		{token.PRINT_INT, "print_int"},
		{token.IDENT, "#0"},

		{token.STORE, "store"},
		{token.IDENT, "#1"},
		{token.COMMA, ","},
		{token.STRING, "steve"},
		{token.PRINT_STR, "print_str"},
		{token.IDENT, "#1"},

		{token.EOF, ""},
	}
	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong, expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - Literal wrong, expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestUnicodeLexer(t *testing.T) {
	input := `世界`
	l := New(input)
	tok := l.NextToken()
	if tok.Type != token.IDENT {
		t.Fatalf("token type wrong, expected=%q, got=%q", token.IDENT, tok.Type)
	}
	if tok.Literal != "世界" {
		t.Fatalf("token literal wrong, expected=%q, got=%q", "世界", tok.Literal)
	}
}

func TestSimpleComment(t *testing.T) {
	input := `# This is a comment
# This is still a comment
print_int #3
# This is a final
print_int #21
# comment on two-lines`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.PRINT_INT, "print_int"},
		{token.IDENT, "#3"},
		{token.PRINT_INT, "print_int"},
		{token.IDENT, "#21"},
		{token.EOF, ""},
	}
	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong, expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - Literal wrong, expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestSimpleJump(t *testing.T) {
	input := `

        jmp exit_here
        store #1, "Can't Happen\n"
        print_str #1
:exit_here
        nop
        exit

`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.JMP, "jmp"},
		{token.IDENT, "exit_here"},

		{token.STORE, "store"},
		{token.IDENT, "#1"},
		{token.COMMA, ","},
		{token.STRING, "Can't Happen\n"},

		{token.PRINT_STR, "print_str"},
		{token.IDENT, "#1"},

		{token.LABEL, ":exit_here"},
		{token.NOP, "nop"},
		{token.EXIT, "exit"},
		{token.EOF, ""},
	}
	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong, expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - Literal wrong, expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}
