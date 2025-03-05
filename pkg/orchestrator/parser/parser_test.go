package parser

import (
	"fmt"
	"testing"
)

func TestTokenize(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
		err      error
	}{
		{"2 + 2", []string{"2", "+", "2"}, nil},
		{"3.14 * -5", []string{"3.14", "*", "-", "5"}, nil},
		{"2 + (3 * 4)", []string{"2", "+", "(", "3", "*", "4", ")"}, nil},
		{"-10 + 20", []string{"-10", "+", "20"}, nil},
		{"10.5 / 2.5", []string{"10.5", "/", "2.5"}, nil},

		{"2 + a", nil, fmt.Errorf("invalid character: a")},
		{"2 # 3", nil, fmt.Errorf("invalid character: #")},

		{"", []string{}, nil},

		{"(2 + 3) * (4 - 1)", []string{"(", "2", "+", "3", ")", "*", "(", "4", "-", "1", ")"}, nil},
		{"-3.14 + 2.71", []string{"-3.14", "+", "2.71"}, nil},
	}

	for _, test := range tests {
		tokens, err := Tokenize(test.input)
		if err != nil && test.err == nil {
			t.Errorf("Tokenize(%q) returned unexpected error: %v", test.input, err)
		}
		if err == nil && test.err != nil {
			t.Errorf("Tokenize(%q) expected error: %v", test.input, test.err)
		}
		if err != nil && test.err != nil && err.Error() != test.err.Error() {
			t.Errorf("Tokenize(%q) returned wrong error: got %v want %v", test.input, err, test.err)
		}
		if !compareSlices(tokens, test.expected) {
			t.Errorf("Tokenize(%q) = %v, expected %v", test.input, tokens, test.expected)
		}
	}
}

func TestToPostfix(t *testing.T) {
	tests := []struct {
		input    []string
		expected []string
		err      error
	}{
		{[]string{"2", "+", "2"}, []string{"2", "2", "+"}, nil},
		{[]string{"3", "*", "(", "4", "+", "5", ")"}, []string{"3", "4", "5", "+", "*"}, nil},
		{[]string{"(", "2", "+", "3", ")", "*", "4"}, []string{"2", "3", "+", "4", "*"}, nil},
		{[]string{"2", "+", "3", "*", "4"}, []string{"2", "3", "4", "*", "+"}, nil},
		{[]string{"10", "/", "2", "+", "5"}, []string{"10", "2", "/", "5", "+"}, nil},

		{[]string{"2", "+", "(", "3", "*", "4"}, nil, fmt.Errorf("mismatched parentheses")},

		{[]string{"2", "+", "a"}, nil, fmt.Errorf("unsudnsud")},

		{[]string{}, []string{}, nil},
	}

	for _, test := range tests {
		result, err := toPostfix(test.input)
		if err != nil && test.err == nil {
			t.Errorf("toPostfix(%v) returned unexpected error: %v", test.input, err)
		}
		if err == nil && test.err != nil {
			t.Errorf("toPostfix(%v) expected error: %v", test.input, test.err)
		}
		if err != nil && test.err != nil && err.Error() != test.err.Error() {
			t.Errorf("toPostfix(%v) returned wrong error: got %v want %v", test.input, err, test.err)
		}
		if !compareSlices(result, test.expected) {
			t.Errorf("toPostfix(%v) = %v, expected %v", test.input, result, test.expected)
		}
	}
}

func compareSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
