package evaluator

import (
	"testing"

	"github.com/lqqyt2423/go-monkey/lexer"
	"github.com/lqqyt2423/go-monkey/object"
	"github.com/lqqyt2423/go-monkey/parser"
)

func TestEvalIntegerExpression(t *testing.T) {
	tests := []struct {
		input     string
		wantValue int64
	}{
		{"5", 5},
		{"10", 10},
		{"-5", -5},
		{"-10", -10},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()
			obj := Eval(program)
			iobj, ok := obj.(*object.Integer)
			if !ok {
				t.Fatalf("should be *object.Integer, but got %T", obj)
			}
			if iobj.Value != tt.wantValue {
				t.Fatalf("value want %d, but got %d", tt.wantValue, iobj.Value)
			}
		})
	}
}

func TestEvalBooleanExpression(t *testing.T) {
	tests := []struct {
		input     string
		wantValue bool
	}{
		{"true", true},
		{"false", false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()
			obj := Eval(program)
			bobj, ok := obj.(*object.Boolean)
			if !ok {
				t.Fatalf("should be *object.Boolean, but got %T", obj)
			}
			if bobj.Value != tt.wantValue {
				t.Fatalf("value want %t, but got %t", tt.wantValue, bobj.Value)
			}
		})
	}
}

func TestBandOperator(t *testing.T) {
	tests := []struct {
		input     string
		wantValue bool
	}{
		{"!true", false},
		{"!false", true},
		{"!5", false},
		{"!!true", true},
		{"!!false", false},
		{"!!5", true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()
			obj := Eval(program)
			bobj, ok := obj.(*object.Boolean)
			if !ok {
				t.Fatalf("should be *object.Boolean, but got %T", obj)
			}
			if bobj.Value != tt.wantValue {
				t.Fatalf("value want %t, but got %t", tt.wantValue, bobj.Value)
			}
		})
	}
}
