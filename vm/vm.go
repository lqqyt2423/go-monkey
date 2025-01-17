package vm

import (
	"fmt"

	"github.com/lqqyt2423/go-monkey/code"
	"github.com/lqqyt2423/go-monkey/compiler"
	"github.com/lqqyt2423/go-monkey/object"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

const (
	MaxFrames   = 1024
	StackSize   = 2048
	GlobalsSize = 65536
)

type VM struct {
	constants []object.Object
	globals   []object.Object

	stack []object.Object
	sp    int

	frames      []*Frame
	framesIndex int
}

func New(bytecode *compiler.ByteCode) *VM {
	mainFn := &object.CompiledFunction{Instructions: bytecode.Instructions}
	mainFrame := NewFrame(mainFn, 0)

	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame

	return &VM{
		constants: bytecode.Constants,
		globals:   make([]object.Object, GlobalsSize),

		stack: make([]object.Object, StackSize),
		sp:    0,

		frames:      frames,
		framesIndex: 1,
	}
}

func NewWithGlobalsStore(bytecode *compiler.ByteCode, s []object.Object) *VM {
	vm := New(bytecode)
	vm.globals = s
	return vm
}

func (vm *VM) StackTop() object.Object {
	if vm.sp == 0 {
		return nil
	}
	return vm.stack[vm.sp-1]
}

func (vm *VM) LastPoppedStackElem() object.Object {
	return vm.stack[vm.sp]
}

func (vm *VM) Run() error {
	var ip int
	var ins code.Instructions
	var op code.Opcode

	for vm.currentFrame().ip < len(vm.currentFrame().Instructions())-1 {
		vm.currentFrame().ip++

		ip = vm.currentFrame().ip
		ins = vm.currentFrame().Instructions()
		op = code.Opcode(ins[ip])

		switch op {
		case code.OpConstant:
			constIndex := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			err := vm.push(vm.constants[constIndex])
			if err != nil {
				return err
			}
		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv:
			vm.execBinaryOperation(op)
		case code.OpPop:
			vm.pop()
		case code.OpTrue:
			vm.push(TRUE)
		case code.OpFalse:
			vm.push(FALSE)
		case code.OpEqual, code.OpNotEqual, code.OpGreaterThan:
			vm.execCompareOperation(op)
		case code.OpMinus:
			val := vm.pop()
			if val.Type() != object.INTEGER_OBJ {
				return fmt.Errorf("unsupported type for minus operation: %s", val.Type())
			}
			v := val.(*object.Integer).Value
			vm.push(&object.Integer{Value: -v})
		case code.OpBang:
			val := vm.pop()
			if isTruthy(val) {
				vm.push(FALSE)
			} else {
				vm.push(TRUE)
			}
		case code.OpJump:
			pos := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip = pos - 1
		case code.OpJumpNotTruthy:
			pos := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2
			condition := vm.pop()
			if !isTruthy(condition) {
				vm.currentFrame().ip = pos - 1
			}
		case code.OpNull:
			vm.push(NULL)
		case code.OpSetGlobal:
			index := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2
			vm.globals[index] = vm.pop()
		case code.OpGetGlobal:
			index := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2
			vm.push(vm.globals[index])
		case code.OpSetLocal:
			localIndex := int(ins[ip+1])
			vm.currentFrame().ip += 1
			frame := vm.currentFrame()
			vm.stack[frame.basePointer+int(localIndex)] = vm.pop()
		case code.OpGetLocal:
			localIndex := int(ins[ip+1])
			vm.currentFrame().ip += 1
			frame := vm.currentFrame()
			vm.push(vm.stack[frame.basePointer+int(localIndex)])
		case code.OpGetBuiltin:
			builtinIndex := int(ins[ip+1])
			vm.currentFrame().ip += 1
			definition := object.Builtins[builtinIndex]
			vm.push(definition.Builtin)
		case code.OpArray:
			arrLen := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2
			arr := &object.Array{Elements: make([]object.Object, arrLen)}
			for i := arrLen - 1; i >= 0; i-- {
				arr.Elements[i] = vm.pop()
			}
			vm.push(arr)
		case code.OpIndex:
			idx := vm.pop()
			if idx.Type() != object.INTEGER_OBJ {
				return fmt.Errorf("invalid index type %s", idx.Type())
			}
			arr := vm.pop()
			if arr.Type() != object.ARRAY_OBJ {
				return fmt.Errorf("type %s can not index", arr.Type())
			}
			idxVal := idx.(*object.Integer).Value
			arrElements := arr.(*object.Array).Elements
			if idxVal < 0 || idxVal >= int64(len(arrElements)) {
				vm.push(NULL)
			} else {
				vm.push(arrElements[idxVal])
			}
		case code.OpCall:
			numArgs := int(ins[ip+1])
			vm.currentFrame().ip += 1

			callee := vm.stack[vm.sp-numArgs-1]
			switch callee := callee.(type) {
			case *object.CompiledFunction:
				fn := callee
				if numArgs != fn.NumParameters {
					return fmt.Errorf("wrong number of arguments: want=%d, got=%d", fn.NumParameters, numArgs)
				}
				frame := NewFrame(fn, vm.sp-numArgs)
				vm.pushFrame(frame)
				vm.sp = frame.basePointer + fn.NumLocals
			case *object.Builtin:
				builtin := callee
				args := vm.stack[vm.sp-numArgs : vm.sp]
				result := builtin.Fn(args...)
				vm.sp = vm.sp - numArgs - 1
				if result != nil {
					vm.push(result)
				} else {
					vm.push(NULL)
				}
			default:
				return fmt.Errorf("calling non-function")
			}

		case code.OpReturnValue:
			returnValue := vm.pop()
			frame := vm.popFrame()
			vm.sp = frame.basePointer - 1
			vm.push(returnValue)
		case code.OpReturn:
			frame := vm.popFrame()
			vm.sp = frame.basePointer - 1
			vm.push(NULL)
		}
	}
	return nil
}

func (vm *VM) execBinaryOperation(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()
	rightType := right.Type()
	leftType := left.Type()
	if rightType == object.INTEGER_OBJ && leftType == object.INTEGER_OBJ {
		return vm.execBinaryIntegerOperation(op, left, right)
	}
	if rightType == object.STRING_OBJ && leftType == object.STRING_OBJ {
		return vm.execBinaryStringOperation(op, left, right)
	}
	return fmt.Errorf("unsupported types for binary operation: %s %s", leftType, rightType)
}

func (vm *VM) execBinaryIntegerOperation(op code.Opcode, left, right object.Object) error {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value
	var result int64
	switch op {
	case code.OpAdd:
		result = leftVal + rightVal
	case code.OpSub:
		result = leftVal - rightVal
	case code.OpMul:
		result = leftVal * rightVal
	case code.OpDiv:
		result = leftVal / rightVal
	default:
		return fmt.Errorf("invalid op %v", op)
	}
	vm.push(&object.Integer{Value: result})
	return nil
}

func (vm *VM) execBinaryStringOperation(op code.Opcode, left, right object.Object) error {
	if op != code.OpAdd {
		return fmt.Errorf("unknown string operator: %v", op)
	}
	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value
	vm.push(&object.String{Value: leftVal + rightVal})
	return nil
}

func (vm *VM) execCompareOperation(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()
	rightType := right.Type()
	leftType := left.Type()

	if rightType == object.INTEGER_OBJ && leftType == object.INTEGER_OBJ {
		leftVal := left.(*object.Integer).Value
		rightVal := right.(*object.Integer).Value
		switch op {
		case code.OpGreaterThan:
			if leftVal > rightVal {
				vm.push(nativeBoolToBooleanObject(true))
			} else {
				vm.push(nativeBoolToBooleanObject(false))
			}
		case code.OpEqual:
			if leftVal == rightVal {
				vm.push(nativeBoolToBooleanObject(true))
			} else {
				vm.push(nativeBoolToBooleanObject(false))
			}
		case code.OpNotEqual:
			if leftVal != rightVal {
				vm.push(nativeBoolToBooleanObject(true))
			} else {
				vm.push(nativeBoolToBooleanObject(false))
			}
		default:
			return fmt.Errorf("invalid op %v", op)
		}
		return nil
	}

	if rightType == object.BOOLEAN_OBJ && leftType == object.BOOLEAN_OBJ {
		leftVal := left.(*object.Boolean).Value
		rightVal := right.(*object.Boolean).Value
		switch op {
		case code.OpEqual:
			if leftVal == rightVal {
				vm.push(nativeBoolToBooleanObject(true))
			} else {
				vm.push(nativeBoolToBooleanObject(false))
			}
		case code.OpNotEqual:
			if leftVal != rightVal {
				vm.push(nativeBoolToBooleanObject(true))
			} else {
				vm.push(nativeBoolToBooleanObject(false))
			}
		default:
			return fmt.Errorf("invalid op %v", op)
		}
		return nil
	}

	return fmt.Errorf("unsupported types for compare operation: %s %v %s", leftType, op, rightType)
}

func (vm *VM) push(o object.Object) error {
	if vm.sp >= StackSize {
		return fmt.Errorf("stack overflow")
	}
	vm.stack[vm.sp] = o
	vm.sp++
	return nil
}

func (vm *VM) pop() object.Object {
	o := vm.stack[vm.sp-1]
	vm.sp--
	return o
}

func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.framesIndex-1]
}

func (vm *VM) pushFrame(f *Frame) {
	vm.frames[vm.framesIndex] = f
	vm.framesIndex++
}

func (vm *VM) popFrame() *Frame {
	vm.framesIndex--
	return vm.frames[vm.framesIndex]
}

func nativeBoolToBooleanObject(b bool) *object.Boolean {
	if b {
		return TRUE
	}
	return FALSE
}

func isTruthy(obj object.Object) bool {
	switch obj := obj.(type) {
	case *object.Boolean:
		return obj.Value
	case *object.Null:
		return false
	default:
		return true
	}
}
