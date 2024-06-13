package parser

import (
	"testing"

	"github.com/lqqyt2423/go-monkey/ast"
	"github.com/lqqyt2423/go-monkey/lexer"
)

func TestLetStatements(t *testing.T) {
	input := `
let x = 5;
let y = 10;
let foobar = 838383;
`
	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)
	if program == nil {
		t.Fatalf("ParseProgram should not be nil")
	}
	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements len want %d, but got %d", 3, len(program.Statements))
	}

	tests := []struct {
		wantIdentifier string
		wantStr        string
	}{
		{"x", "let x = 5;"},
		{"y", "let y = 10;"},
		{"foobar", "let foobar = 838383;"},
	}

	for i, tt := range tests {
		stmt := program.Statements[i]
		if !testLetStatement(t, stmt, tt.wantIdentifier) {
			return
		}
		if stmt.String() != tt.wantStr {
			t.Fatalf("stmt.String() want %q, but got %q", tt.wantStr, stmt.String())
		}
	}
}

func TestReturnStatements(t *testing.T) {
	input := `
return 5;
return 5 * 5;
return x(a);
`
	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)
	if program == nil {
		t.Fatalf("ParseProgram should not be nil")
	}
	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements len want %d, but got %d", 3, len(program.Statements))
	}

	wantStr := []string{
		"return 5;",
		"return (5 * 5);",
		"return x(a) ;",
	}

	for i, stmt := range program.Statements {
		if stmt.TokenLiteral() != "return" {
			t.Errorf("stmt.TokenLiteral() want %q, but got %q", "return", stmt.TokenLiteral())
		}
		if _, ok := stmt.(*ast.ReturnStatement); !ok {
			t.Errorf("stmt not *ast.ReturnStatement, got %T", stmt)
		}
		if stmt.String() != wantStr[i] {
			t.Errorf("stmt.String() want %q, but got %q", wantStr[i], stmt.String())
		}
	}
}

func TestString(t *testing.T) {
	input := "let a = b;"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if input != program.String() {
		t.Fatalf("input should equal out string, but got %s", program.String())
	}
}

func TestIdentifierExpression(t *testing.T) {
	input := "x;"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements len want %d, but got %d", 1, len(program.Statements))
	}
	stmt := program.Statements[0]
	exStmt, ok := stmt.(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("should be *ast.ExpressionStatement, but got %T", stmt)
	}
	exp, ok := exStmt.Expression.(*ast.Identifier)
	if !ok {
		t.Fatalf("should be *ast.Identifier, but got %T", exStmt.Expression)
	}
	if exp.Value != "x" {
		t.Fatalf("want x, but got %q", exp.Value)
	}
}

func TestIntegerLiteralExpression(t *testing.T) {
	input := "5;"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements len want %d, but got %d", 1, len(program.Statements))
	}
	stmt := program.Statements[0]
	exStmt, ok := stmt.(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("should be *ast.ExpressionStatement, but got %T", stmt)
	}
	exp, ok := exStmt.Expression.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("should be *ast.IntegerLiteral, but got %T", exStmt.Expression)
	}
	if exp.Value != 5 {
		t.Fatalf("want 5, but got %d", exp.Value)
	}
	if exp.TokenLiteral() != "5" {
		t.Fatalf("want 5, but got %s", exp.TokenLiteral())
	}
}

func TestPrefixExpression(t *testing.T) {
	tests := []struct {
		input            string
		wantOperator     string
		wantIntegerValue int64
	}{
		{"!5;", "!", 5},
		{"-15", "-", 15},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)
			if len(program.Statements) != 1 {
				t.Fatalf("program.Statements len should be 1, but got %d", len(program.Statements))
			}
			stmt := program.Statements[0]
			exStmt, ok := stmt.(*ast.ExpressionStatement)
			if !ok {
				t.Fatalf("should be *ast.ExpressionStatement, but got %T", stmt)
			}
			pe, ok := exStmt.Expression.(*ast.PrefixExpression)
			if !ok {
				t.Fatalf("should be *ast.PrefixExpression, but got %T", exStmt)
			}
			if pe.Operator != tt.wantOperator {
				t.Fatalf("Operator want %s, but got %s", tt.wantOperator, pe.Operator)
			}
			ilExp, ok := pe.Right.(*ast.IntegerLiteral)
			if !ok {
				t.Fatalf("should be *ast.IntegerLiteral, but got %T", pe.Right)
			}
			if ilExp.Value != tt.wantIntegerValue {
				t.Fatalf("want %d, but got %d", tt.wantIntegerValue, ilExp.Value)
			}
		})
	}
}

func TestInfixExpression(t *testing.T) {
	tests := []struct {
		input        string
		wantLeft     int64
		wantOperator string
		wantRight    int64
	}{
		{"1+2;", 1, "+", 2},
		{"3-4;", 3, "-", 4},
		{"5*6;", 5, "*", 6},
		{"7/8;", 7, "/", 8},
		{"11<12;", 11, "<", 12},
		{"13>14;", 13, ">", 14},
		{"15==16;", 15, "==", 16},
		{"17!=18;", 17, "!=", 18},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)
			if len(program.Statements) != 1 {
				t.Fatalf("program.Statements len should be 1, but got %d", len(program.Statements))
			}
			stmt := program.Statements[0]
			exStmt, ok := stmt.(*ast.ExpressionStatement)
			if !ok {
				t.Fatalf("should be *ast.ExpressionStatement, but got %T", stmt)
			}
			ie, ok := exStmt.Expression.(*ast.InfixExpression)
			if !ok {
				t.Fatalf("should be *ast.InfixExpression, but got %T", exStmt)
			}
			if ie.Operator != tt.wantOperator {
				t.Fatalf("Operator want %s, but got %s", tt.wantOperator, ie.Operator)
			}
			ilExp, ok := ie.Left.(*ast.IntegerLiteral)
			if !ok {
				t.Fatalf("left should be *ast.IntegerLiteral, but got %T", ie.Left)
			}
			if ilExp.Value != tt.wantLeft {
				t.Fatalf("left want %d, but got %d", tt.wantLeft, ilExp.Value)
			}
			ilExp, ok = ie.Right.(*ast.IntegerLiteral)
			if !ok {
				t.Fatalf("right should be *ast.IntegerLiteral, but got %T", ie.Right)
			}
			if ilExp.Value != tt.wantRight {
				t.Fatalf("right want %d, but got %d", tt.wantRight, ilExp.Value)
			}
		})
	}
}

func TestIfExpression(t *testing.T) {
	tests := []struct {
		input   string
		wantStr string
	}{
		{
			input:   "if (x == true) { 1 } else { 2 }",
			wantStr: "if (x == true) 1 else 2",
		},
		{
			input:   "if (y) { x }",
			wantStr: "if y x",
		},
		{
			input:   "if (x) { x; y; z; }",
			wantStr: "if x xyz",
		},
		{
			input:   "if (x) { x; y; z; 1 }",
			wantStr: "if x xyz1",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)
			if len(program.Statements) != 1 {
				t.Fatalf("program.Statements len should be 1, but got %d", len(program.Statements))
			}
			stmt := program.Statements[0]
			exStmt, ok := stmt.(*ast.ExpressionStatement)
			if !ok {
				t.Fatalf("should be *ast.ExpressionStatement, but got %T", stmt)
			}
			_, ok = exStmt.Expression.(*ast.IfExpression)
			if !ok {
				t.Fatalf("should be *ast.IfExpression, but got %T", exStmt)
			}
			if program.String() != tt.wantStr {
				t.Fatalf("program.String() want %q, but got %q", tt.wantStr, program.String())
			}
		})
	}
}

func TestFunctionLiteral(t *testing.T) {
	tests := []struct {
		input string
	}{
		{
			input: "fn(x, y) { x + y; }",
		},
		{
			input: "fn() {}",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)
			if len(program.Statements) != 1 {
				t.Fatalf("program.Statements len should be 1, but got %d", len(program.Statements))
			}
			stmt := program.Statements[0]
			exStmt, ok := stmt.(*ast.ExpressionStatement)
			if !ok {
				t.Fatalf("should be *ast.ExpressionStatement, but got %T", stmt)
			}
			_, ok = exStmt.Expression.(*ast.FunctionLiteral)
			if !ok {
				t.Fatalf("should be *ast.FunctionLiteral, but got %T", exStmt)
			}
		})
	}
}

func TestCallExpression(t *testing.T) {
	tests := []struct {
		input   string
		wantStr string
	}{
		{
			input:   "call()",
			wantStr: "call() ",
		},
		{
			input:   "call(x)",
			wantStr: "call(x) ",
		},
		{
			input:   "call(x, y)",
			wantStr: "call(x, y) ",
		},
		{
			input:   "xyz(1, 2 * 3, 4 / 5)",
			wantStr: "xyz(1, (2 * 3), (4 / 5)) ",
		},
		{
			"add(a, b, 1, 2 * 3, 4 + 5, add(6, 7 * 8))",
			"add(a, b, 1, (2 * 3), (4 + 5), add(6, (7 * 8)) ) ",
		},
		{
			"add(a + b + c * d / f + g)",
			"add((((a + b) + ((c * d) / f)) + g)) ",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)
			if len(program.Statements) != 1 {
				t.Fatalf("program.Statements len should be 1, but got %d", len(program.Statements))
			}
			stmt := program.Statements[0]
			exStmt, ok := stmt.(*ast.ExpressionStatement)
			if !ok {
				t.Fatalf("should be *ast.ExpressionStatement, but got %T", stmt)
			}
			_, ok = exStmt.Expression.(*ast.CallExpression)
			if !ok {
				t.Fatalf("should be *ast.CallExpression, but got %T", exStmt)
			}
			if program.String() != tt.wantStr {
				t.Fatalf("program.String() want %q, but got %q", tt.wantStr, program.String())
			}
		})
	}
}

func TestComplexExpression(t *testing.T) {
	tests := []struct {
		input   string
		wantStr string
	}{
		{"1+2+3;", "((1 + 2) + 3)"},
		{"1+2-3;", "((1 + 2) - 3)"},
		{"1*2*3;", "((1 * 2) * 3)"},
		{"1*2/3;", "((1 * 2) / 3)"},
		{"1*2+3;", "((1 * 2) + 3)"},
		{"1/2-3;", "((1 / 2) - 3)"},
		{"1+2*3;", "(1 + (2 * 3))"},
		{"1-2/3;", "(1 - (2 / 3))"},
		{"-1+2-3;", "(((-1) + 2) - 3)"},
		{"-1+2*3;", "((-1) + (2 * 3))"},
		{"1+2*-3;", "(1 + (2 * (-3)))"},
		{
			"-a * b",
			"((-a) * b)",
		},
		{
			"!-a",
			"(!(-a))",
		},
		{
			"a + b + c",
			"((a + b) + c)",
		},
		{
			"a + b - c",
			"((a + b) - c)",
		},
		{
			"a * b * c",
			"((a * b) * c)",
		},
		{
			"a * b / c",
			"((a * b) / c)",
		},
		{
			"a + b / c",
			"(a + (b / c))",
		},
		{
			"a + b * c + d / e - f",
			"(((a + (b * c)) + (d / e)) - f)",
		},
		{
			"3 + 4; -5 * 5",
			"(3 + 4)((-5) * 5)",
		},
		{
			"5 > 4 == 3 < 4",
			"((5 > 4) == (3 < 4))",
		},
		{
			"5 < 4 != 3 > 4",
			"((5 < 4) != (3 > 4))",
		},
		{
			"3 + 4 * 5 == 3 * 1 + 4 * 5",
			"((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))",
		},
		{"true == false", "(true == false)"},
		{
			"true",
			"true",
		},
		{
			"false",
			"false",
		},
		{
			"3 > 5 == false",
			"((3 > 5) == false)",
		},
		{
			"3 < 5 == true",
			"((3 < 5) == true)",
		},
		{"1+(2+3);", "(1 + (2 + 3))"},
		{
			"1 + (2 + 3) + 4",
			"((1 + (2 + 3)) + 4)",
		},
		{
			"(5 + 5) * 2",
			"((5 + 5) * 2)",
		},
		{
			"2 / (5 + 5)",
			"(2 / (5 + 5))",
		},
		{
			"-(5 + 5)",
			"(-(5 + 5))",
		},
		{
			"!(true == true)",
			"(!(true == true))",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)
			if program.String() != tt.wantStr {
				t.Fatalf("program.String() want %q, but got %q", tt.wantStr, program.String())
			}
		})
	}
}

func checkParserErrors(t *testing.T, p *Parser) {
	if len(p.Errors()) == 0 {
		return
	}
	t.Errorf("parser had %d errors", len(p.Errors()))
	for _, msg := range p.Errors() {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}

func testLetStatement(t *testing.T, s ast.Statement, name string) bool {
	if s.TokenLiteral() != "let" {
		t.Errorf("s.TokenLiteral() want %q, but got %q", "let", s.TokenLiteral())
		return false
	}
	letStmt, ok := s.(*ast.LetStatement)
	if !ok {
		t.Errorf("s not *ast.LetStatement, got %T", s)
		return false
	}
	if letStmt.Name.Value != name {
		t.Errorf("letStmt.Name.Value want %q, but got %q", name, letStmt.Name.Value)
		return false
	}
	if letStmt.Name.TokenLiteral() != name {
		t.Errorf("letStmt.Name.TokenLiteral() want %q, but got %q", name, letStmt.Name.TokenLiteral())
		return false
	}
	return true
}
