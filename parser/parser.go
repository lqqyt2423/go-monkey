package parser

import (
	"fmt"
	"strconv"

	"github.com/lqqyt2423/go-monkey/ast"
	"github.com/lqqyt2423/go-monkey/lexer"
	"github.com/lqqyt2423/go-monkey/token"
)

type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken  token.Token
	peekToken token.Token

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:              l,
		prefixParseFns: make(map[token.TokenType]prefixParseFn),
		infixParseFns:  make(map[token.TokenType]infixParseFn),
	}

	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)
	p.registerPrefix(token.LPAREN, p.parseGroupExpression)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)
	p.registerPrefix(token.LBRACKET, p.parseArrayLiteral)
	p.registerPrefix(token.LBRACE, p.parseHashLiteral)

	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	p.registerInfix(token.LBRACKET, p.parseIndexExpression)

	p.nextToken()
	p.nextToken()
	return p
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

type Precedence int

const (
	LOWEST      Precedence = iota + 1
	EQUALS                 // ==
	LESSGREATER            // > or <
	SUM                    // +
	PRODUCT                // *
	PREFIX                 // -X or !X
	CALL                   // myFunction(X)
	INDEX                  // array[index]
)

var precedences = map[token.TokenType]Precedence{
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.ASTERISK: PRODUCT,
	token.SLASH:    PRODUCT,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LPAREN:   CALL,
	token.LBRACKET: INDEX,
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	for p.curToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}
	return program
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{
		Token: p.curToken,
	}
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.Name = &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}
	if !p.expectPeek(token.ASSIGN) {
		return nil
	}
	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{
		Token: p.curToken,
	}
	p.nextToken()
	stmt.ReturnValue = p.parseExpression(LOWEST)
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{
		Token: p.curToken,
	}
	stmt.Expression = p.parseExpression(LOWEST)
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	stmt := &ast.BlockStatement{
		Token: p.curToken,
	}
	p.nextToken()
	for p.curToken.Type != token.RBRACE && p.curToken.Type != token.EOF {
		s := p.parseStatement()
		if s != nil {
			stmt.Statements = append(stmt.Statements, s)
		}
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseExpression(precedence Precedence) ast.Expression {
	prefixFn, ok := p.prefixParseFns[p.curToken.Type]
	if !ok {
		p.errors = append(p.errors, fmt.Sprintf("no prefix fn found %s", p.curToken.Type))
		return nil
	}
	exp := prefixFn()

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infixFn, ok := p.infixParseFns[p.peekToken.Type]
		if !ok {
			return exp
		}
		p.nextToken()
		exp = infixFn(exp)
	}

	return exp
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	v, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		p.errors = append(p.errors, fmt.Sprintf("could not parse %q as integer", p.curToken.Literal))
		return nil
	}
	return &ast.IntegerLiteral{
		Token: p.curToken,
		Value: v,
	}
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}
}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{
		Token: p.curToken,
		Value: p.curToken.Type == token.TRUE,
	}
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	exp := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}
	p.nextToken()
	exp.Right = p.parseExpression(PREFIX)
	return exp
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	exp := &ast.InfixExpression{
		Token:    p.curToken,
		Left:     left,
		Operator: p.curToken.Literal,
	}
	precedences := p.curPrecedence()
	p.nextToken()
	exp.Right = p.parseExpression(precedences)
	return exp
}

func (p *Parser) parseGroupExpression() ast.Expression {
	p.nextToken()
	exp := p.parseExpression(LOWEST)
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return exp
}

func (p *Parser) parseIfExpression() ast.Expression {
	exp := &ast.IfExpression{
		Token: p.curToken,
	}
	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	exp.Condition = p.parseExpression(LOWEST)
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	exp.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(token.ELSE) {
		p.nextToken()
		if !p.expectPeek(token.LBRACE) {
			return nil
		}
		exp.Alternative = p.parseBlockStatement()
	}

	return exp
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	exp := &ast.FunctionLiteral{
		Token: p.curToken,
	}
	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	if !p.peekTokenIs(token.RPAREN) {
		if !p.expectPeek(token.IDENT) {
			return nil
		}
		param := p.parseIdentifier().(*ast.Identifier)
		exp.Parameters = append(exp.Parameters, param)
		for p.peekTokenIs(token.COMMA) {
			p.nextToken()
			if !p.expectPeek(token.IDENT) {
				return nil
			}
			param := p.parseIdentifier().(*ast.Identifier)
			exp.Parameters = append(exp.Parameters, param)
		}
	}
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	exp.Body = p.parseBlockStatement()
	return exp
}

func (p *Parser) parseCallExpression(left ast.Expression) ast.Expression {
	exp := &ast.CallExpression{
		Token:    p.curToken,
		Function: left,
	}
	if !p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		argExp := p.parseExpression(LOWEST)
		exp.Arguments = append(exp.Arguments, argExp)
		for p.peekTokenIs(token.COMMA) {
			p.nextToken()
			p.nextToken()
			argExp := p.parseExpression(LOWEST)
			exp.Arguments = append(exp.Arguments, argExp)
		}
	}
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return exp
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	exp := &ast.ArrayLiteral{Token: p.curToken}
	if !p.peekTokenIs(token.RBRACKET) {
		p.nextToken()
		ele := p.parseExpression(LOWEST)
		exp.Elements = append(exp.Elements, ele)
		for p.peekTokenIs(token.COMMA) {
			p.nextToken()
			p.nextToken()
			ele := p.parseExpression(LOWEST)
			exp.Elements = append(exp.Elements, ele)
		}
	}
	if !p.expectPeek(token.RBRACKET) {
		return nil
	}
	return exp
}

func (p *Parser) parseHashLiteral() ast.Expression {
	exp := &ast.HashLiteral{
		Token: p.curToken,
		Pairs: make(map[ast.Expression]ast.Expression),
	}
	if !p.peekTokenIs(token.RBRACE) {
		p.nextToken()
		key := p.parseExpression(LOWEST)
		if !p.expectPeek(token.COLON) {
			return nil
		}
		p.nextToken()
		value := p.parseExpression(LOWEST)
		exp.Pairs[key] = value
		for p.peekTokenIs(token.COMMA) {
			p.nextToken()
			p.nextToken()
			key := p.parseExpression(LOWEST)
			if !p.expectPeek(token.COLON) {
				return nil
			}
			p.nextToken()
			value := p.parseExpression(LOWEST)
			exp.Pairs[key] = value
		}
	}
	if !p.expectPeek(token.RBRACE) {
		return nil
	}
	return exp
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{
		Token: p.curToken,
		Left:  left,
	}
	p.nextToken()
	exp.Index = p.parseExpression(LOWEST)
	if !p.expectPeek(token.RBRACKET) {
		return nil
	}
	return exp
}

func (p *Parser) curTokenIs(typ token.TokenType) bool {
	return p.curToken.Type == typ
}

func (p *Parser) peekTokenIs(typ token.TokenType) bool {
	return p.peekToken.Type == typ
}

func (p *Parser) expectPeek(typ token.TokenType) bool {
	if p.peekTokenIs(typ) {
		p.nextToken()
		return true
	} else {
		p.peekError(typ)
		return false
	}
}

func (p *Parser) peekError(typ token.TokenType) {
	msg := fmt.Sprintf("expect next token to be %s, but got %s", typ, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) registerPrefix(typ token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[typ] = fn
}

func (p *Parser) registerInfix(typ token.TokenType, fn infixParseFn) {
	p.infixParseFns[typ] = fn
}

func (p *Parser) curPrecedence() Precedence {
	return precedences[p.curToken.Type]
}

func (p *Parser) peekPrecedence() Precedence {
	return precedences[p.peekToken.Type]
}
