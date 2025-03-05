package calculation

import (
	"go/parser"
	"testing"
)

func TestCalculate(t *testing.T) {
	tests := []struct {
		name      string
		arg1      string
		arg2      string
		operation string
		want      float64
	}{
		{"Addition", "2", "3", "+", 5},
		{"Subtraction", "5", "3", "-", 2},
		{"Multiplication", "4", "3", "*", 12},
		{"Division", "10", "2", "/", 5},
		{"Division by zero", "5", "0", "/", 0},
		{"Invalid operation", "2", "2", "%", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Calculate(tt.arg1, tt.arg2, tt.operation)
			if got != tt.want {
				t.Errorf("Calculate(%s, %s, %s) = %v, want %v", tt.arg1, tt.arg2, tt.operation, got, tt.want)
			}
		})
	}
}

func TestEvaluateAST(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		want       float64
	}{
		{"Simple addition", "2 + 3", 5},
		{"Complex expression", "2 * 3 + 4 / 2", 8},
		{"Parentheses", "(2 + 3) * 4", 20},
		{"Negative numbers", "-5 + 3", -2},
		{"Float numbers", "3.5 * 2", 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree, err := parser.ParseExpr(tt.expression)
			if err != nil {
				t.Fatalf("Failed to parse expression: %v", err)
			}
			got := evaluateAST(tree)
			if got != tt.want {
				t.Errorf("evaluateAST(%s) = %v, want %v", tt.expression, got, tt.want)
			}
		})
	}
}
