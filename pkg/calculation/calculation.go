package calculation

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
)

func Calculate(arg1, arg2, operation string) float64 {
	expr := arg1 + operation + arg2
	tree, err := parser.ParseExpr(expr)
	if err != nil {
		return 0
	}

	result := evaluateAST(tree)
	return result
}

func evaluateAST(node ast.Node) float64 {
	switch n := node.(type) {
	case *ast.BinaryExpr:
		x := evaluateAST(n.X)
		y := evaluateAST(n.Y)
		switch n.Op {
		case token.ADD:
			return x + y
		case token.SUB:
			return x - y
		case token.MUL:
			return x * y
		case token.QUO:
			if y != 0 {
				return x / y
			}
			return 0
		}
	case *ast.BasicLit:
		value, _ := strconv.ParseFloat(n.Value, 64)
		return value
	}
	return 0
}
