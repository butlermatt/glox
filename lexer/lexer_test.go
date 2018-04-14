package lexer

import "testing"

func TestLexer_ScanTokensCount(t *testing.T) {
	// Should be a token for each token in the input, plus one for EOF
	tests := []struct {
		input string
		count int
	}{
		{`+-=[]()/{}!`, 12},
		{`+
// Ignore comment
-
{}[]`, 7},
		{`>=,==,<=,!= // Test 2 character tokens`, 8},
		{`~`, 2},           // Illegal and EOF
		{`"something"`, 2}, // String
		{`"something`, 2},  // Unterminated String
		{"`A fancy\nMultiline\nString`", 2},
		{`2342.2323`, 2},
	}

	for i, tt := range tests {
		l := New(tt.input)
		l.ScanTokens()

		if len(l.tokens) != tt.count {
			t.Errorf("test %d: did not generate enough tokens. expected=%d, got=%d", i+1, tt.count, len(l.tokens))
			t.Errorf("%+v\n", l.tokens)
		}
	}
}

func TestLexer_NextToken(t *testing.T) {
	input := `+-
{}[]
123.456
// Ignore this
!= == <= >=
"a string"
`
	input += "`A\nMultiline\nString`\n"
	input += `!
,
.
fun something true false
if and or else for while
class null this super return;
`

	expected := []struct {
		ty      TokenType
		literal string
		value   interface{}
		line    int
	}{
		{Plus, "+", nil, 1},
		{Minus, "-", nil, 1},
		{LBrace, "{", nil, 2},
		{RBrace, "}", nil, 2},
		{LBracket, "[", nil, 2},
		{RBracket, "]", nil, 2},
		{Number, "123.456", float64(123.456), 3},
		{BangEq, "!=", nil, 5},
		{EqualEq, "==", nil, 5},
		{LessEq, "<=", nil, 5},
		{GreaterEq, ">=", nil, 5},
		{String, `"a string"`, "a string", 6},
		{String, "`A\nMultiline\nString`", `A
Multiline
String`, 7},
		{Bang, "!", nil, 10},
		{Comma, ",", nil, 11},
		{Dot, ".", nil, 12},
		{Fun, "fun", nil, 13},
		{Ident, "something", nil, 13},
		{True, "true", nil, 13},
		{False, "false", nil, 13},
		{If, "if", nil, 14},
		{And, "and", nil, 14},
		{Or, "or", nil, 14},
		{Else, "else", nil, 14},
		{For, "for", nil, 14},
		{While, "while", nil, 14},
		{Class, "class", nil, 15},
		{Null, "null", nil, 15},
		{This, "this", nil, 15},
		{Super, "super", nil, 15},
		{Return, "return", nil, 15},
		{Semicolon, ";", nil, 15},
		{EOF, "", nil, 16},
	}

	l := New(input)
	l.ScanTokens()

	for i, expect := range expected {
		tok := l.NextToken()
		if tok == nil {
			t.Fatalf("test %d: unexpected missing token. expected=%q", i, expect.ty)
		}

		if tok.Type != expect.ty {
			t.Errorf("test %d: unexpected token. expected=%q, got=%q", i, expect.ty, tok.Type)
		}

		if tok.Lexeme != expect.literal {
			t.Errorf("test %d: unexpected lexeme. expected=%q, got=%q", i, expect.literal, tok.Lexeme)
		}

		if tok.Literal != expect.value {
			t.Errorf("test %d: unexpected literal value. expected=%v, got=%v", i, expect.value, tok.Literal)
		}

		if tok.Line != expect.line {
			t.Errorf("test %d: unexpected line. expected=%d, got=%d", i, expect.line, tok.Line)
		}
	}
}
