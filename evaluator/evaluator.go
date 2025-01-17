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

var builtins = map[string]*object.Builtin{
	"len": object.GetBuiltinByName("len"),
}

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
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}
	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value)
	case *ast.ArrayLiteral:
		var elements []object.Object
		for _, ele := range node.Elements {
			element := Eval(ele, env)
			if isError(element) {
				return element
			}
			elements = append(elements, element)
		}
		return &object.Array{Elements: elements}
	case *ast.HashLiteral:
		hash := &object.Hash{
			Pairs: make(map[string]object.Object),
		}
		for k, v := range node.Pairs {
			key := Eval(k, env)
			if isError(key) {
				return key
			}
			keyStr, ok := key.(*object.String)
			if !ok {
				return newError("hash key should be String, but got %s", key.Type())
			}
			val := Eval(v, env)
			if isError(val) {
				return val
			}
			hash.Pairs[keyStr.Value] = val
		}
		return hash
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
		if ok {
			return val
		}
		val, ok = builtins[node.Value]
		if ok {
			return val
		}
		return newError("identifier not found: %s", node.Value)
	case *ast.FunctionLiteral:
		return &object.Function{
			Parameters: node.Parameters,
			Body:       node.Body,
			Env:        env,
		}
	case *ast.CallExpression:
		return evalCallExpression(node, env)
	case *ast.IndexExpression:
		return evalIndexExpression(node, env)
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
	if left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ {
		leftVal := left.(*object.String).Value
		rightVal := right.(*object.String).Value
		switch operator {
		case "+":
			return &object.String{Value: leftVal + rightVal}
		case "==":
			return &object.Boolean{Value: leftVal == rightVal}
		case "!=":
			return &object.Boolean{Value: leftVal != rightVal}
		}
		return newError("type mismatch: %s %s %s", left.Type(), operator, right.Type())
	}

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

	var args []object.Object
	for _, argument := range node.Arguments {
		arg := Eval(argument, env)
		if isError(arg) {
			return arg
		}
		args = append(args, arg)
	}

	switch funcObj := function.(type) {
	case *object.Builtin:
		return funcObj.Fn(args...)
	case *object.Function:
		if len(funcObj.Parameters) != len(args) {
			return newError("arguments len %d mismatch, want %d", len(args), len(funcObj.Parameters))
		}

		callEnv := object.ExtendEnvironment(funcObj.Parameters, args, funcObj.Env)
		val := Eval(funcObj.Body, callEnv)
		if rval, ok := val.(*object.ReturnValue); ok {
			return rval.Value
		}
		return val
	default:
		return newError("type mismatch: %s()", function.Type())
	}
}

func evalIndexExpression(node *ast.IndexExpression, env *object.Environment) object.Object {
	left := Eval(node.Left, env)
	if isError(left) {
		return left
	}
	index := Eval(node.Index, env)
	if isError(index) {
		return index
	}

	switch leftObj := left.(type) {
	case *object.Array:
		indexVal, ok := index.(*object.Integer)
		if !ok {
			return newError("type mismatch: %s", index.Type())
		}
		if indexVal.Value < 0 || indexVal.Value >= int64(len(leftObj.Elements)) {
			return newError("out of index")
		}
		return leftObj.Elements[indexVal.Value]
	case *object.Hash:
		indexVal, ok := index.(*object.String)
		if !ok {
			return newError("type mismatch: %s", index.Type())
		}
		val, ok := leftObj.Pairs[indexVal.Value]
		if ok {
			return val
		} else {
			return NULL
		}
	default:
		return newError("type mismatch: %s", left.Type())
	}
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
