package lexer

import (
	"testing"

	"github.com/lqqyt2423/go-monkey/token"
)

func TestNextTokenSimple(t *testing.T) {
	input := `=+(){},;-*/<>!
true false if else return
== !=
"hello"
"world"
1
`
	tests := []struct {
		wantType    token.TokenType
		wantLiteral string
	}{
		{token.ASSIGN, "="},
		{token.PLUS, "+"},
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RBRACE, "}"},
		{token.COMMA, ","},
		{token.SEMICOLON, ";"},
		{token.MINUS, "-"},
		{token.ASTERISK, "*"},
		{token.SLASH, "/"},
		{token.LT, "<"},
		{token.GT, ">"},
		{token.BANG, "!"},
		{token.TRUE, "true"},
		{token.FALSE, "false"},
		{token.IF, "if"},
		{token.ELSE, "else"},
		{token.RETURN, "return"},
		{token.EQ, "=="},
		{token.NOT_EQ, "!="},
		{token.STRING, "hello"},
		{token.STRING, "world"},
		{token.INT, "1"},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.wantType {
			t.Fatalf("tests[%d] - tokentype wrong. want=%q, got=%q", i, tt.wantType, tok.Type)
		}
		if tok.Literal != tt.wantLiteral {
			t.Fatalf("tests[%d] - literal wrong. want=%q, got=%q", i, tt.wantLiteral, tok.Literal)
		}
	}
}

func TestNextToken(t *testing.T) {
	input := `let five = 5;
let ten = 10;

let add = fn(x, y) {
  x + y;
};

let result = add(five, ten);
"hello" 123`

	tests := []struct {
		wantType    token.TokenType
		wantLiteral string
	}{
		{token.LET, "let"},
		{token.IDENT, "five"},
		{token.ASSIGN, "="},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.LET, "let"},
		{token.IDENT, "ten"},
		{token.ASSIGN, "="},
		{token.INT, "10"},
		{token.SEMICOLON, ";"},
		{token.LET, "let"},
		{token.IDENT, "add"},
		{token.ASSIGN, "="},
		{token.FUNCTION, "fn"},
		{token.LPAREN, "("},
		{token.IDENT, "x"},
		{token.COMMA, ","},
		{token.IDENT, "y"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.IDENT, "x"},
		{token.PLUS, "+"},
		{token.IDENT, "y"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.SEMICOLON, ";"},
		{token.LET, "let"},
		{token.IDENT, "result"},
		{token.ASSIGN, "="},
		{token.IDENT, "add"},
		{token.LPAREN, "("},
		{token.IDENT, "five"},
		{token.COMMA, ","},
		{token.IDENT, "ten"},
		{token.RPAREN, ")"},
		{token.SEMICOLON, ";"},
		{token.STRING, "hello"},
		{token.INT, "123"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.wantType {
			t.Fatalf("tests[%d] - tokentype wrong. want=%q, got=%q", i, tt.wantType, tok.Type)
		}
		if tok.Literal != tt.wantLiteral {
			t.Fatalf("tests[%d] - literal wrong. want=%q, got=%q", i, tt.wantLiteral, tok.Literal)
		}
	}
}
