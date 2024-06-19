package object

import (
	"fmt"
	"strings"

	"github.com/lqqyt2423/go-monkey/ast"
)

type Environment struct {
	store map[string]Object
	outer *Environment
}

func NewEnvironment() *Environment {
	return &Environment{
		store: make(map[string]Object),
	}
}

func ExtendEnvironment(variables []*ast.Identifier, values []Object, baseEnv *Environment) *Environment {
	store := make(map[string]Object)
	for i, key := range variables {
		store[key.Value] = values[i]
	}
	return &Environment{
		store: store,
		outer: baseEnv,
	}
}

func (env *Environment) Set(key string, val Object) Object {
	env.store[key] = val
	return val
}

func (env *Environment) Get(key string) (Object, bool) {
	obj, ok := env.store[key]
	if !ok && env.outer != nil {
		return env.outer.Get(key)
	}
	return obj, ok
}

type ObjectType string

type Object interface {
	Type() ObjectType
	Inspect() string
}

const (
	INTEGER_OBJ      ObjectType = "INTEGER"
	STRING_OBJ       ObjectType = "STRING"
	BOOLEAN_OBJ      ObjectType = "BOOLEAN"
	NULL_OBJ         ObjectType = "NULL"
	RETURN_VALUE_OBJ ObjectType = "RETURN_VALUE"
	ERROR_OBJ        ObjectType = "ERROR"
	FUNCTION_OBJ     ObjectType = "FUNCTION"
	BUILTIN_OBJ      ObjectType = "BUILTIN"
)

type Integer struct {
	Value int64
}

func (i *Integer) Type() ObjectType { return INTEGER_OBJ }
func (i *Integer) Inspect() string  { return fmt.Sprintf("%d", i.Value) }

type String struct {
	Value string
}

func (s *String) Type() ObjectType { return STRING_OBJ }
func (s *String) Inspect() string  { return `"` + s.Value + `"` }

type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (b *Boolean) Inspect() string  { return fmt.Sprintf("%t", b.Value) }

type Null struct{}

func (n *Null) Type() ObjectType { return NULL_OBJ }
func (n *Null) Inspect() string  { return "null" }

type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }
func (rv *ReturnValue) Inspect() string  { return rv.Value.Inspect() }

type Error struct {
	Message string
}

func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) Inspect() string  { return "ERROR: " + e.Message }

type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
}

func (f *Function) Type() ObjectType { return FUNCTION_OBJ }
func (f *Function) Inspect() string {
	var out strings.Builder
	var params []string
	for _, param := range f.Parameters {
		params = append(params, param.String())
	}
	out.WriteString("fn(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") {\n")
	out.WriteString(f.Body.String())
	out.WriteString("\n}")
	return out.String()
}

type BuiltinFunction func(args ...Object) Object

type Builtin struct {
	Fn BuiltinFunction
}

func (b *Builtin) Type() ObjectType { return BUILTIN_OBJ }
func (b *Builtin) Inspect() string  { return "builtin function" }
