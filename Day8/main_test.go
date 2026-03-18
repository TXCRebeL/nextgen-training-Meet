package main

import (
	"errors"
	"math"
	"testing"
)

func TestBracketMatcher(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		wantErr error
	}{
		{"Empty", "", nil},
		{"Single pair", "()", nil},
		{"Types", "()[]{}<>", nil},
		{"Nested", "(<[{}]>)", nil},
		{"Unclosed", "(", &MismatchError{}},
		{"Extra close", ")", &MismatchError{}},
		{"Mismatched", "(]", &MismatchError{}},
		{"Wrong matching", "([)]", &MismatchError{}},
		{"Multiple mismatch", "((", &MismatchError{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := BracketMatcher(tt.expr)
			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("BracketMatcher(%q) expected error, got nil", tt.expr)
					return
				}
				var mismatchErr *MismatchError
				if !errors.As(err, &mismatchErr) {
					t.Errorf("BracketMatcher(%q) error = %T, want %T", tt.expr, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("BracketMatcher(%q) unexpected error: %v", tt.expr, err)
			}
		})
	}
}

func TestInfixToPostfix(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		want    string
		wantErr error
	}{
		{"Simple add", "2 + 3", "2 3 +", nil},
		{"Float add", "2.4 + 3.6", "2.4 3.6 +", nil},
		{"Precedence", "2 + 3 * 4", "2 3 4 * +", nil},
		{"Brackets", "(2 + 3) * 4", "2 3 + 4 *", nil},
		{"Power", "2 ^ 3", "2 3 ^", nil},
		{"Unary minus start", "-2 + 3", "2 u- 3 +", nil},
		{"Unary minus bracket", "5 * (-3 + 2)", "5 3 u- 2 + *", nil},
		{"Multiple unary", "--2", "2 u- u-", nil},
		{"Power right-assoc", "2 ^ 3 ^ 2", "2 3 2 ^ ^", nil},
		{"Empty", "", "", nil},
		{"Invalid char", "2 @ 3", "", &SyntaxError{}},
		{"Missing closing", "(2 + 3", "", &MismatchError{}},
		{"Missing opening", "2 + 3)", "", &MismatchError{}},
		{"Operator at end", "2 +", "2 +", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := InfixToPostfix(tt.expr)
			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("InfixToPostfix(%q) expected error, got nil", tt.expr)
					return
				}
				var syntaxErr *SyntaxError
				var mismatchErr *MismatchError
				if errors.As(tt.wantErr, &syntaxErr) && !errors.As(err, &syntaxErr) {
					t.Errorf("InfixToPostfix(%q) error = %T, want SyntaxError", tt.expr, err)
				}
				if errors.As(tt.wantErr, &mismatchErr) && !errors.As(err, &mismatchErr) {
					t.Errorf("InfixToPostfix(%q) error = %T, want MismatchError", tt.expr, err)
				}
				return
			}
			if err != nil {
				t.Errorf("InfixToPostfix(%q) unexpected error: %v", tt.expr, err)
				return
			}
			if got != tt.want {
				t.Errorf("InfixToPostfix(%q) = %q, want %q", tt.expr, got, tt.want)
			}
		})
	}
}

func TestPostfixEvaluation(t *testing.T) {
	tests := []struct {
		name    string
		postfix string
		want    float64
		wantErr error
	}{
		{"Simple add", "2 3 +", 5, nil},
		{"Float add", "2.4 3.6 +", 6.0, nil},
		{"Sub", "10 4 -", 6, nil},
		{"Mul", "3 4 *", 12, nil},
		{"Div", "10 2 /", 5, nil},
		{"Pow", "2 3 ^", 8, nil},
		{"Unary", "2 u- 3 +", 1, nil},
		{"Div by zero", "10 0 /", 0, DivisionByZeroError},
		{"Malformed too many", "2 3 4 +", 0, &SyntaxError{}},
		{"Malformed not enough", "2 +", 0, &SyntaxError{}},
		{"Malformed unary", "u-", 0, &SyntaxError{}},
		{"Invalid operand", "abc", 0, &SyntaxError{}},
		{"Empty", "", 0, nil},
		{"Whitespace", "   ", 0, nil},
		{"Zero only", "0", 0, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PostfixEvaluation(tt.postfix)
			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("PostfixEvaluation(%q) expected error, got nil", tt.postfix)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					var syntaxErr *SyntaxError
					if errors.As(err, &syntaxErr) && !errors.As(tt.wantErr, &syntaxErr) {
						t.Errorf("PostfixEvaluation(%q) error = %T, want SyntaxError", tt.postfix, err)
					}
					return
				}
				return
			}
			if err != nil {
				t.Errorf("PostfixEvaluation(%q) unexpected error: %v", tt.postfix, err)
				return
			}
			if math.Abs(got-tt.want) > 0.000001 {
				t.Errorf("PostfixEvaluation(%q) = %v, want %v", tt.postfix, got, tt.want)
			}
		})
	}
}

func TestGetPrecedenceForOp(t *testing.T) {
	if GetPrecedenceForOp("u-") != 4 {
		t.Error("u- precedence should be 4")
	}
	if GetPrecedenceForOp("^") != 3 {
		t.Error("^ precedence should be 3")
	}
	if GetPrecedenceForOp("*") != 2 {
		t.Error("* precedence should be 2")
	}
	if GetPrecedenceForOp("+") != 1 {
		t.Error("+ precedence should be 1")
	}
	if GetPrecedenceForOp("invalid") != 0 {
		t.Error("invalid op precedence should be 0")
	}
}

func TestErrors(t *testing.T) {
	mErr := &MismatchError{Position: 5, Message: "test"}
	if mErr.Error() != "Mismatch Error at position 5: test" {
		t.Errorf("MismatchError.Error() = %q", mErr.Error())
	}
	sErr := &SyntaxError{Position: 10, Message: "syntax"}
	if sErr.Error() != "Syntax Error at position 10: syntax" {
		t.Errorf("SyntaxError.Error() = %q", sErr.Error())
	}
}
