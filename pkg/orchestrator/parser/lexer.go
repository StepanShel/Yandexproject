package parser

import (
	"fmt"
	"strconv"
	"unicode"
)

func Tokenize(expression string) ([]string, error) {
	var tokens []string
	var buffer []rune

	for i, char := range expression {
		if unicode.IsSpace(char) {
			continue
		}

		if unicode.IsDigit(char) || char == '.' {
			buffer = append(buffer, char)
			continue
		}

		if char == '-' && (i == 0 || isOperator(string(expression[i-1]))) {
			buffer = append(buffer, char)
			continue
		}

		if isOperator(string(char)) || char == '(' || char == ')' {
			if len(buffer) > 0 {
				tokens = append(tokens, string(buffer))
				buffer = []rune{}
			}
			tokens = append(tokens, string(char))
			continue
		}

		return nil, fmt.Errorf("invalid character: %v", string(char))
	}

	if len(buffer) > 0 {
		tokens = append(tokens, string(buffer))
	}

	return tokens, nil
}

func isOperator(s string) bool {
	return s == "+" || s == "-" || s == "*" || s == "/"
}

func isNumber(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}
