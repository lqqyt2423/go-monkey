package evaluator

import (
	"fmt"

	"github.com/lqqyt2423/go-monkey/ast"
	"github.com/lqqyt2423/go-monkey/object"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(node, env)
	case *ast.ReturnStatement:
		v := Eval(node.ReturnValue, env)
		if isError(v) {
			return v
		}
		return &object.ReturnValue{Value: v}
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case *ast.BlockStatement:
		return evalStatements(node.Statements, env)
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value)
	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)
	case *ast.IfExpression:
		return evalIfExpression(node, env)
	case *ast.LetStatement:
		value := Eval(node.Value, env)
		if isError(value) {
			return value
		}
		env.Set(node.Name.Value, value)
		return NULL
	case *ast.Identifier:
		val, ok := env.Get(node.Value)
		if !ok {
			return newError("identifier not found: %s", node.Value)
		}
		return val
	case *ast.FunctionLiteral:
		return &object.Function{
			Parameters: node.Parameters,
			Body:       node.Body,
			Env:        env,
		}
	case *ast.CallExpression:
		return evalCallExpression(node, env)
	default:
		return NULL
	}
}

func evalProgram(node *ast.Program, env *object.Environment) object.Object {
	result := evalStatements(node.Statements, env)
	switch result := result.(type) {
	case *object.ReturnValue:
		return result.Value
	default:
		return result
	}
}

func evalStatements(stmts []ast.Statement, env *object.Environment) object.Object {
	var result object.Object
	for _, stmt := range stmts {
		result = Eval(stmt, env)
		if result.Type() == object.RETURN_VALUE_OBJ || isError(result) {
			return result
		}
	}
	return result
}

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusOperatorExpression(right)
	default:
		return NULL
	}
}

func evalBangOperatorExpression(right object.Object) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

func evalMinusOperatorExpression(right object.Object) object.Object {
	if right.Type() != object.INTEGER_OBJ {
		return newError("type mismatch: -%s", right.Type())
	}
	value := right.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

func evalInfixExpression(operator string, left object.Object, right object.Object) object.Object {
	if left.Type() != object.INTEGER_OBJ || right.Type() != object.INTEGER_OBJ {
		if operator == "==" || operator == "!=" {
			return evalEqInfixCompress(operator, left, right)
		}
		return newError("type mismatch: %s + %s", left.Type(), right.Type())
	}
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value
	switch operator {
	case "+":
		return &object.Integer{Value: leftValue + rightValue}
	case "-":
		return &object.Integer{Value: leftValue - rightValue}
	case "*":
		return &object.Integer{Value: leftValue * rightValue}
	case "/":
		return &object.Integer{Value: leftValue / rightValue}
	case "<":
		if leftValue < rightValue {
			return TRUE
		} else {
			return FALSE
		}
	case ">":
		if leftValue > rightValue {
			return TRUE
		} else {
			return FALSE
		}
	case "==":
		if leftValue == rightValue {
			return TRUE
		} else {
			return FALSE
		}
	case "!=":
		if leftValue != rightValue {
			return TRUE
		} else {
			return FALSE
		}
	default:
		return NULL
	}
}

func evalEqInfixCompress(operator string, left object.Object, right object.Object) object.Object {
	switch operator {
	case "==":
		return nativeBoolToBooleanObject(left == right)
	case "!=":
		return nativeBoolToBooleanObject(left != right)
	default:
		return NULL
	}
}

func evalIfExpression(node *ast.IfExpression, env *object.Environment) object.Object {
	condition := Eval(node.Condition, env)
	if isError(condition) {
		return condition
	}
	if isTruthy(condition) {
		return Eval(node.Consequence, env)
	}
	if node.Alternative != nil {
		return Eval(node.Alternative, env)
	}
	return NULL
}

func evalCallExpression(node *ast.CallExpression, env *object.Environment) object.Object {
	function := Eval(node.Function, env)
	if isError(function) {
		return function
	}
	funcObj, ok := function.(*object.Function)
	if !ok {
		return newError("type mismatch: %s()", function.Type())
	}

	var args []object.Object
	for _, argument := range node.Arguments {
		arg := Eval(argument, env)
		if isError(arg) {
			return arg
		}
		args = append(args, arg)
	}

	if len(funcObj.Parameters) != len(args) {
		return newError("arguments len %d mismatch, want %d", len(args), len(funcObj.Parameters))
	}

	callEnv := object.ExtendEnvironment(funcObj.Parameters, args, funcObj.Env)
	val := Eval(funcObj.Body, callEnv)
	if rval, ok := val.(*object.ReturnValue); ok {
		return rval.Value
	}
	return val
}

func nativeBoolToBooleanObject(b bool) *object.Boolean {
	if b {
		return TRUE
	}
	return FALSE
}

func isTruthy(v object.Object) bool {
	switch v {
	case FALSE:
		return false
	case NULL:
		return false
	default:
		return true
	}
}

func newError(format string, a ...any) object.Object {
	return &object.Error{
		Message: fmt.Sprintf(format, a...),
	}
}

func isError(v object.Object) bool {
	_, ok := v.(*object.Error)
	return ok
}
