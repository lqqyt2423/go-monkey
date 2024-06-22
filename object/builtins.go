package object

import (
	"fmt"
)

var Builtins = []struct {
	Name    string
	Builtin *Builtin
}{
	{
		"len",
		&Builtin{
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return newError("arguments len %d mismatch, want %d", len(args), 1)
				}
				arg := args[0]
				switch argObj := arg.(type) {
				case *String:
					return &Integer{Value: int64(len(argObj.Value))}
				case *Array:
					return &Integer{Value: int64(len(argObj.Elements))}
				case *Hash:
					return &Integer{Value: int64(len(argObj.Pairs))}
				default:
					return newError("type mismatch: %s", arg.Type())
				}
			},
		},
	},
	{
		"puts",
		&Builtin{
			Fn: func(args ...Object) Object {
				for _, arg := range args {
					fmt.Println(arg.Inspect())
				}
				return nil
			},
		},
	},
}

func GetBuiltinByName(name string) *Builtin {
	for _, def := range Builtins {
		if def.Name == name {
			return def.Builtin
		}
	}
	return nil
}

func newError(format string, a ...any) Object {
	return &Error{
		Message: fmt.Sprintf(format, a...),
	}
}
