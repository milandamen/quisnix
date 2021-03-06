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
	for _, decl := range declarations {
		switch d := decl.(type) {
		case *parser.FunctionDeclaration:
			err := p.addFunctionDeclaration(d)
			if err != nil {
				return errors.Wrapf(err, "cannot print function '%s'", d.Name)
			}
		default:
			return errors.New("unknown declaration type")
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

func (p *LLVMPrinter) addFunctionDeclaration(decl *parser.FunctionDeclaration) error {
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

	f := p.module.NewFunc(decl.Name, retType, params...)
	b := f.NewBlock("")

	scope, err := getFuncVariableScope(decl.FunctionDefinition.FunctionType.Parameters, params)
	if err != nil {
		return err
	}

	if err := p.addStatements(f, b, decl.FunctionDefinition.Statements, scope); err != nil {
		return err
	}

	// When to allocate on the heap instead of the stack:
	//  1. When the lifetime of the value exceeds the current function
	//  2. When the value can grow (arrays), put the whole struct on the heap, and return pointer to value (like append() in Go)

	return nil
}

func (p *LLVMPrinter) addStatements(f *ir.Func, b *ir.Block, statements []parser.Statement, outsideScopeVars map[*parser.VariableDeclaration]value.Value) (map[*parser.VariableDeclaration]value.Value, error) {
	overwrittenVars := make(map[*parser.VariableDeclaration]value.Value)
	scope := make(map[*parser.VariableDeclaration]value.Value)

	for _, statement := range statements {
		if stmtHVarDecl, ok := statement.(parser.StatementHavingVariableDeclaration); ok {
			decl := stmtHVarDecl.GetVariableDeclaration()
			varDecl, ok2 := decl.(*parser.VariableDeclaration)
			if !ok2 {
				return nil, errors.New("compiler error: statement having declaration is not a variable declaration")
			}
			varVal, inScope := scope[varDecl]
			if !inScope {
				var ok3 bool
				varVal, ok3 = overwrittenVars[varDecl]
				if !ok3 {
					var ok4 bool
					varVal, ok4 = outsideScopeVars[varDecl]
					if !ok4 {
						return nil, errors.New("compiler error: variable declaration not in scope")
					}
				}
			}

			var newVal value.Value
			var vals []value.Value
			var err error
			switch s := statement.(type) {
			case *parser.AssignStatement:
				vals, err = p.getExpressionValues(b, s.Expression)
				if err != nil {
					return nil, err
				}
				if len(vals) != 1 {
					return nil, errors.New("compiler error: resulting expression values must have len 1")
				}

				newVal = vals[0]
			case *parser.AddAssignStatement:
				vals, err = p.getExpressionValues(b, s.Expression)
				if err != nil {
					return nil, err
				}
				if len(vals) != 1 {
					return nil, errors.New("compiler error: resulting expression values must have len 1")
				}

				newVal = b.NewAdd(varVal, vals[0])
			case *parser.SubtractAssignStatement:
				vals, err = p.getExpressionValues(b, s.Expression)
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
				vals, err := p.getExpressionValues(b, stmt.ReturnExpressions[0])
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
		} else {
			return nil, errors.New("compiler error: unsupported statement")
		}
	}

	return overwrittenVars, nil
}

func (p *LLVMPrinter) getExpressionValues(b *ir.Block, expression parser.Expression) ([]value.Value, error) {

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
