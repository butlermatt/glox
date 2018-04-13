package lexer

import "testing"

func TestLexer_ScanTokensCount(t *testing.T) {
	tests := []struct {
		input string
		count int
	}{
		{`+-=[](){}!`, 10},
		{`+
// Ignore comment
-`, 2},
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
