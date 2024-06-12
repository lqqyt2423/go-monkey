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
	}{
		{"x"},
		{"y"},
		{"foobar"},
	}

	for i, tt := range tests {
		stmt := program.Statements[i]
		if !testLetStatement(t, stmt, tt.wantIdentifier) {
			return
		}
	}
}

func TestReturnStatements(t *testing.T) {
	input := `
return 5;
return 5 * 5;
return fn(a);
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

	for _, stmt := range program.Statements {
		if stmt.TokenLiteral() != "return" {
			t.Errorf("stmt.TokenLiteral() want %q, but got %q", "return", stmt.TokenLiteral())
		}
		if _, ok := stmt.(*ast.ReturnStatement); !ok {
			t.Errorf("stmt not *ast.ReturnStatement, got %T", stmt)
		}
	}
}

func TestString(t *testing.T) {
	t.Skip()
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
