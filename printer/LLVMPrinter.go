package printer

import (
	"fmt"
	"io"

	"github.com/llir/llvm/ir/value"

	"github.com/pkg/errors"

	"github.com/milandamen/quisnix/parser"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/types"
)

type LLVMPrinter struct {
	module *ir.Module
}

func (p *LLVMPrinter) Print(w io.Writer, declarations []parser.Declaration) error {
	p.module = ir.NewModule()
	funcList := make(map[*parser.FunctionDeclaration]*ir.Func)
	for _, decl := range declarations {
		switch d := decl.(type) {
		case *parser.FunctionDeclaration:
			err := p.addFunctionDeclaration(d, funcList)
			if err != nil {
				return errors.Wrapf(err, "cannot print function '%s'", d.Name)
			}
		default:
			return errors.New("unknown declaration type")
		}
	}

	for funcDecl, f := range funcList {
		err := p.addFunctionStatements(funcDecl, f, funcList)
		if err != nil {
			return errors.Wrapf(err, "cannot print function '%s'", funcDecl.Name)
		}
	}

	//i32 := types.I32
	//g2 := constant.NewInt(i32, 3)
	//m := ir.NewModule()
	//
	//param := ir.NewParam("x", i32)
	//f := m.NewFunc("main", i32, param)
	//entry := f.NewBlock("")
	//tmp1 := entry.NewMul(param, g2)
	//entry.NewRet(tmp1)
	//
	//a := f.NewBlock("")
	//tmp2 := a.NewMul(param, g2)
	//a.NewRet(tmp2)

	_, err := p.module.WriteTo(w)
	return err
}

func (p *LLVMPrinter) addFunctionDeclaration(decl *parser.FunctionDeclaration, funcList map[*parser.FunctionDeclaration]*ir.Func) error {
	retTypeFields := decl.FunctionDefinition.FunctionType.ReturnTypes
	var retTypes []types.Type
	for _, f := range retTypeFields {
		typ, err := getLLVMType(f.VariableDeclaration.TypeDeclaration.Type)
		if err != nil {
			return errors.Wrap(err, "cannot not get LLVM type for return type")
		}

		retTypes = append(retTypes, typ)
	}

	var retType types.Type
	if len(retTypes) == 1 {
		retType = retTypes[0]
	} else {
		retType = types.Void
	}

	params, err := getLLVMFunctionParams(decl.FunctionDefinition.FunctionType.Parameters, retTypes)
	if err != nil {
		return err
	}

	machineName := decl.MachineName
	if decl.Name == "main" {
		machineName = "main" // TODO detect via annotation if this function should be main.
	}
	f := p.module.NewFunc(machineName, retType, params...)
	funcList[decl] = f
	return nil
}

func (p *LLVMPrinter) addFunctionStatements(decl *parser.FunctionDeclaration, f *ir.Func, funcList map[*parser.FunctionDeclaration]*ir.Func) error {
	b := f.NewBlock("")

	variableScope, err := getFuncVariableScope(decl.FunctionDefinition.FunctionType.Parameters, f.Params)
	if err != nil {
		return err
	}

	overwrittenVars, err := p.addStatements(b, decl.FunctionDefinition.Statements, variableScope, funcList)
	if err != nil {
		return err
	}

	_ = overwrittenVars // TODO use PHI with overwritten vars

	// When to allocate on the heap instead of the stack:
	//  1. When the lifetime of the value exceeds the current function
	//  2. When the value can grow (arrays), put the whole struct on the heap, and return pointer to value (like append() in Go)

	return nil
}

func (p *LLVMPrinter) addStatements(b *ir.Block,
	statements []parser.Statement,
	outsideScopeVars map[*parser.VariableDeclaration]value.Value,
	funcList map[*parser.FunctionDeclaration]*ir.Func) (map[*parser.VariableDeclaration]value.Value, error) {

	overwrittenVars := make(map[*parser.VariableDeclaration]value.Value)
	scope := make(map[*parser.VariableDeclaration]value.Value)
	// TODO use PHI with overwritten vars

	for _, statement := range statements {
		if stmtHVarDecl, ok := statement.(parser.StatementHavingVariableDeclaration); ok {
			decl := stmtHVarDecl.GetVariableDeclaration()
			varDecl, ok2 := decl.(*parser.VariableDeclaration)
			if !ok2 {
				return nil, errors.New("compiler error: statement having declaration is not a variable declaration")
			}
			varVal, inScope, err := p.getScopeVariableValue(varDecl, scope, overwrittenVars, outsideScopeVars)
			if err != nil {
				return nil, err
			}

			var newVal value.Value
			var vals []value.Value
			switch s := statement.(type) {
			case *parser.AssignStatement:
				vals, err = p.getExpressionValues(b, s.Expression, scope, overwrittenVars, outsideScopeVars, funcList)
				if err != nil {
					return nil, err
				}
				if len(vals) != 1 {
					return nil, errors.New("compiler error: resulting expression values must have len 1")
				}

				newVal = vals[0]
			case *parser.AddAssignStatement:
				vals, err = p.getExpressionValues(b, s.Expression, scope, overwrittenVars, outsideScopeVars, funcList)
				if err != nil {
					return nil, err
				}
				if len(vals) != 1 {
					return nil, errors.New("compiler error: resulting expression values must have len 1")
				}

				newVal = b.NewAdd(varVal, vals[0])
			case *parser.SubtractAssignStatement:
				vals, err = p.getExpressionValues(b, s.Expression, scope, overwrittenVars, outsideScopeVars, funcList)
				if err != nil {
					return nil, err
				}
				if len(vals) != 1 {
					return nil, errors.New("compiler error: resulting expression values must have len 1")
				}

				newVal = b.NewSub(varVal, vals[0])
			case *parser.IncrementStatement:
				newVal = b.NewAdd(varVal, constant.NewInt(types.I32, 1))
			case *parser.DecrementStatement:
				newVal = b.NewSub(varVal, constant.NewInt(types.I32, 1))
			default:
				return nil, errors.New("compiler error: unknown statement")
			}

			if inScope {
				scope[varDecl] = newVal
			} else {
				overwrittenVars[varDecl] = newVal
			}
		} else if stmt, ok := statement.(*parser.ReturnStatement); ok {
			if len(stmt.ReturnExpressions) == 1 {
				vals, err := p.getExpressionValues(b, stmt.ReturnExpressions[0], scope, overwrittenVars, outsideScopeVars, funcList)
				if err != nil {
					return nil, err
				}
				if len(vals) != 1 {
					return nil, errors.New("compiler error: resulting expression values must have len 1")
				}

				b.NewRet(vals[0])
			} else {
				return nil, errors.New("multiple return values not yet supported")
			}
		} else if stmt, ok := statement.(*parser.VariableDeclaration); ok {
			if _, ok2 := stmt.TypeDeclaration.Type.(parser.BasicType); !ok2 {
				return nil, errors.New("declaring a non-basic variable is not yet supported")
			}

			zeroVal, err := p.getZeroValue(stmt.TypeDeclaration.Type)
			if err != nil {
				return nil, errors.Wrap(err, "cannot get zero value for variable")
			}
			scope[stmt] = zeroVal
		} else {
			return nil, errors.New("compiler error: unsupported statement")
		}
	}

	return overwrittenVars, nil
}

func (p *LLVMPrinter) getScopeVariableValue(varDecl *parser.VariableDeclaration, scope, overwrittenVars,
	outsideScopeVars map[*parser.VariableDeclaration]value.Value) (value.Value, bool, error) {

	varVal, inScope := scope[varDecl]
	if !inScope {
		var ok3 bool
		varVal, ok3 = overwrittenVars[varDecl]
		if !ok3 {
			var ok4 bool
			varVal, ok4 = outsideScopeVars[varDecl]
			if !ok4 {
				return nil, false, errors.New("compiler error: variable declaration not in scope")
			}
		}
	}

	return varVal, inScope, nil
}

func (p *LLVMPrinter) getExpressionValues(b *ir.Block, expression parser.Expression, scope, overwrittenVars,
	outsideScopeVars map[*parser.VariableDeclaration]value.Value, funcList map[*parser.FunctionDeclaration]*ir.Func) ([]value.Value, error) {

	switch exp := expression.(type) {
	case *parser.IntegerLiteralExpression:
		val := constant.NewInt(types.I32, int64(exp.Value)) // TODO find out how to use other bit sizes.
		return []value.Value{val}, nil
	case *parser.IdentifierExpression:
		val, _, err := p.getScopeVariableValue(exp.IdentifierDeclaration.(*parser.VariableDeclaration), scope, overwrittenVars, outsideScopeVars)
		if err != nil {
			return nil, err
		}
		return []value.Value{val}, nil
	case *parser.AddExpression:
		val1, err := p.getExpressionValues(b, exp.Left, scope, overwrittenVars, outsideScopeVars, funcList)
		if err != nil {
			return nil, errors.Wrap(err, "cannot 'add' with Left")
		}
		val2, err := p.getExpressionValues(b, exp.Right, scope, overwrittenVars, outsideScopeVars, funcList)
		if err != nil {
			return nil, errors.Wrap(err, "cannot 'add' with Right")
		}

		add := b.NewAdd(val1[0], val2[0]) // TODO what if adding strings?
		return []value.Value{add}, nil
	case *parser.SubtractExpression:
		val1, err := p.getExpressionValues(b, exp.Left, scope, overwrittenVars, outsideScopeVars, funcList)
		if err != nil {
			return nil, errors.Wrap(err, "cannot 'add' with Left")
		}
		val2, err := p.getExpressionValues(b, exp.Right, scope, overwrittenVars, outsideScopeVars, funcList)
		if err != nil {
			return nil, errors.Wrap(err, "cannot 'add' with Right")
		}

		sub := b.NewSub(val1[0], val2[0])
		return []value.Value{sub}, nil
	case *parser.MultiplyExpression:
		val1, err := p.getExpressionValues(b, exp.Left, scope, overwrittenVars, outsideScopeVars, funcList)
		if err != nil {
			return nil, errors.Wrap(err, "cannot 'add' with Left")
		}
		val2, err := p.getExpressionValues(b, exp.Right, scope, overwrittenVars, outsideScopeVars, funcList)
		if err != nil {
			return nil, errors.Wrap(err, "cannot 'add' with Right")
		}

		mul := b.NewMul(val1[0], val2[0])
		return []value.Value{mul}, nil
	case *parser.DivideExpression:
		val1, err := p.getExpressionValues(b, exp.Left, scope, overwrittenVars, outsideScopeVars, funcList)
		if err != nil {
			return nil, errors.Wrap(err, "cannot 'add' with Left")
		}
		val2, err := p.getExpressionValues(b, exp.Right, scope, overwrittenVars, outsideScopeVars, funcList)
		if err != nil {
			return nil, errors.Wrap(err, "cannot 'add' with Right")
		}

		mul := b.NewSDiv(val1[0], val2[0]) // TODO division by zero causes undefined behavior, so code must assert error
		return []value.Value{mul}, nil
	case *parser.FunctionCallExpression:
		switch idExp := exp.CallSource.(type) {
		case *parser.IdentifierExpression:
			switch funcDecl := idExp.IdentifierDeclaration.(type) {
			case *parser.FunctionDeclaration:
				f, ok := funcList[funcDecl]
				if !ok {
					return nil, errors.New("compiler error: function not found for value of exp.CallSource.IdentifierDeclaration")
				}

				params := make([]value.Value, len(exp.Parameters))
				for i, paramExp := range exp.Parameters {
					val, err := p.getExpressionValues(b, paramExp, scope, overwrittenVars, outsideScopeVars, funcList)
					if err != nil {
						return nil, errors.Wrapf(err, "cannot parse parameter at index %d", i)
					}
					params[i] = val[0] // TODO support multiple return values
				}

				call := b.NewCall(f, params...)
				returnTypes := funcDecl.FunctionDefinition.FunctionType.ReturnTypes
				if len(returnTypes) == 0 {
					call.Typ = types.Void
				} else if len(returnTypes) == 1 {
					switch t := returnTypes[0].VariableDeclaration.TypeDeclaration.Type.(type) {
					case parser.BasicType:
						switch t.DataType {
						case parser.IntDataType:
							call.Typ = types.I32
						default:
							return nil, errors.Errorf("compiler error: basic data type '%d' is not implemented", t.DataType)
						}
					default:
						return nil, errors.New("compiler error: unsupported function return type")
					}
				} else {
					return nil, errors.New("Returning multiple values from a function is not yet supported")
				}

				return []value.Value{call}, nil
			case *parser.VariableDeclaration:
				return nil, errors.New("calling a function in a variable is not yet supported")
			default:
				return nil, errors.New("compiler error: unsupported exp.CallSource.IdentifierDeclaration")
			}
		case *parser.FunctionCallExpression:
			return nil, errors.New("calling a function resulting from the call of a function is not yet supported")
		default:
			return nil, errors.New("compiler error: unsupported exp.CallSource")
		}
	default:
		return nil, errors.New("compiler error: unsupported expression type")
	}
}

func (p *LLVMPrinter) getZeroValue(typ parser.Type) (value.Value, error) {
	switch t := typ.(type) {
	case parser.BasicType:
		switch t.DataType {
		case parser.IntDataType:
			return constant.NewInt(types.I32, 0), nil
		default:
			return nil, errors.Errorf("compiler error: basic data type '%d' is not implemented", t.DataType)
		}
	default:
		return nil, errors.New("type is unsupported")
	}
}

func getLLVMFunctionParams(parameters []*parser.Field, returnTypes []types.Type) ([]*ir.Param, error) {
	var params []*ir.Param
	if len(returnTypes) > 1 {
		for i, t := range returnTypes {
			var pt *types.PointerType
			if ppt, ok := t.(*types.PointerType); ok {
				pt = ppt
			} else {
				pt = types.NewPointer(t)
			}

			params = append(params, ir.NewParam(fmt.Sprintf("qx.mulret.%d", i), pt))
		}
	}

	for _, f := range parameters {
		typ, err := getLLVMType(f.VariableDeclaration.TypeDeclaration.Type)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot parse type of parameter '%s'", f.Name)
		}

		params = append(params, ir.NewParam(f.Name, typ))
	}

	return params, nil
}

func getFuncVariableScope(parameters []*parser.Field, irParams []*ir.Param) (map[*parser.VariableDeclaration]value.Value, error) {
	scope := make(map[*parser.VariableDeclaration]value.Value)
	for i, f := range parameters {
		scope[f.VariableDeclaration] = irParams[i]
	}

	return scope, nil
}

func getLLVMType(typ parser.Type) (types.Type, error) {
	switch t := typ.(type) {
	case parser.BasicType:
		switch t.DataType {
		case parser.IntDataType:
			return types.I32, nil
		default:
			return nil, errors.Errorf("unknown/unsupported data type '%d", t.DataType)
		}
	default:
		return nil, errors.Errorf("unknown/unsupported function return type '%s'", typ.TypeName())
	}
}
