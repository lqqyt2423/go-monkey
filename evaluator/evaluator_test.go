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

func TestEvalIntegerInfixExpression(t *testing.T) {
	tests := []struct {
		input     string
		wantValue int64
	}{
		{"1+2", 3},
		{"1-2", -1},
		{"2-1", 1},
		{"3*4", 12},
		{"4/2", 2},
		{"5/2", 2},
		{"1+2*3", 7},
		{"1-2+2", 1},
		{"2-1/1", 1},
		{"3*4+5", 17},
		{"4/2-2", 0},
		{"5/2+5*2", 12},
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

func TestEvalCompareInfixExpression(t *testing.T) {
	tests := []struct {
		input     string
		wantValue bool
	}{
		{"1<2", true},
		{"2<1", false},
		{"1>2", false},
		{"2>1", true},
		{"1==2", false},
		{"1==1", true},
		{"1!=1", false},
		{"1!=2", true},
		{"true==true", true},
		{"false==false", true},
		{"true==false", false},
		{"true!=true", false},
		{"true!=false", true},
		{"false!=false", false},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 > 2) == true", false},
		{"(1 > 2) == false", true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()
			obj := Eval(program)
			iobj, ok := obj.(*object.Boolean)
			if !ok {
				t.Fatalf("should be *object.Boolean, but got %T", obj)
			}
			if iobj.Value != tt.wantValue {
				t.Fatalf("value want %t, but got %t", tt.wantValue, iobj.Value)
			}
		})
	}
}

func TestIfElseExpressions(t *testing.T) {
	tests := []struct {
		input     string
		wantValue int64
		wantNull  bool
	}{
		{"if (true) { 10 }", 10, false},
		{"if (false) { 10 }", 0, true},
		{"if (1) { 10 }", 10, false},
		{"if (1 < 2) { 10 }", 10, false},
		{"if (1 > 2) { 10 }", 0, true},
		{"if (1 > 2) { 10 } else { 20 }", 20, false},
		{"if (1 < 2) { 10 } else { 20 }", 10, false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()
			obj := Eval(program)

			switch obj := obj.(type) {
			case *object.Integer:
				if obj.Value != tt.wantValue {
					t.Fatalf("*object.Integer want %d, but got %d", tt.wantValue, obj.Value)
				}
			case *object.Null:
				if !tt.wantNull {
					t.Fatalf("want *object.Integer, but got *object.Null")
				}
			default:
				t.Fatalf("type error %T", obj)
			}
		})
	}
}

func TestReturnStatement(t *testing.T) {
	tests := []struct {
		input     string
		wantValue int64
	}{
		{"return 10;", 10},
		{"return 10; 9;", 10},
		{"return 2 * 5; 9;", 10},
		{"9; return 2 * 5; 9;", 10},
		{`if (10 > 1) { if (10 > 1) { return 10; } return 1;}`, 10},
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
