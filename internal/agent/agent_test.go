package agent

import (
	"fmt"
	"testing"
)

func TestCalculate(t *testing.T) {
	tests := []struct {
		operation string
		duration  int
		a, b      float64
		expected  float64
		err       error
	}{
		{"+", 100, 2, 3, 5, nil},
		{"-", 100, 5, 3, 2, nil},
		{"*", 100, 2, 3, 6, nil},
		{"/", 100, 6, 3, 2, nil},
		{"/", 100, 6, 0, 0, fmt.Errorf("division by zero")},
		{"%", 100, 6, 3, 0, fmt.Errorf("invalid operator: %%")},
	}

	for _, test := range tests {
		result, err := Calculate(test.operation, test.duration, test.a, test.b)
		if err != nil && test.err == nil {
			t.Errorf("Calculate(%v, %v, %v, %v) returned unexpected error: %v", test.operation, test.duration, test.a, test.b, err)
		}
		if err == nil && test.err != nil {
			t.Errorf("Calculate(%v, %v, %v, %v) expected error: %v", test.operation, test.duration, test.a, test.b, test.err)
		}
		if result != test.expected {
			t.Errorf("Calculate(%v, %v, %v, %v) = %v, expected %v", test.operation, test.duration, test.a, test.b, result, test.expected)
		}
	}
}
