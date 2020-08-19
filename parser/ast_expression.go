package parser

import "github.com/pkg/errors"

type resultingTypeDeclarations interface {
	ResultingTypeDeclarations() ([]*TypeDeclaration, error)
}

type baseExpression struct {
	nodeSource
	typeDeclarations []*TypeDeclaration
}

type dualInputExpression struct {
	baseExpression
	Left  Expression
	Right Expression
}

type dualInputBoolOutputExpression struct {
	dualInputExpression
}

type IntegerLiteralExpression struct {
	baseExpression
	Value int
}

type CharacterLiteralExpression struct {
	baseExpression
	Value byte
}

type StringLiteralExpression struct {
	baseExpression
	Value string
}

type BooleanLiteralExpression struct {
	baseExpression
	Value bool
}

type IdentifierExpression struct {
	baseExpression
	IdentifierDeclaration Declaration
}

type AddExpression struct {
	dualInputExpression
}

type SubtractExpression struct {
	dualInputExpression
}

type MultiplyExpression struct {
	dualInputExpression
}

type DivideExpression struct {
	dualInputExpression
}

type EqualExpression struct {
	dualInputBoolOutputExpression
}

type NotEqualExpression struct {
	dualInputBoolOutputExpression
}

type LessExpression struct {
	dualInputBoolOutputExpression
}

type LessOrEqualExpression struct {
	dualInputBoolOutputExpression
}

type GreaterExpression struct {
	dualInputBoolOutputExpression
}

type GreaterOrEqualExpression struct {
	dualInputBoolOutputExpression
}

type AndExpression struct {
	dualInputBoolOutputExpression
}

type OrExpression struct {
	dualInputBoolOutputExpression
}

type NotExpression struct {
	baseExpression
	Expression Expression
}

type FunctionCallExpression struct {
	baseExpression
	CallSource Expression // Expression representing a function that can be called.
	Parameters []Expression
}

func newBaseExpression(source nodeSource, d ...*TypeDeclaration) baseExpression {
	return baseExpression{
		nodeSource:       source,
		typeDeclarations: d,
	}
}

func newIntegerLiteralExpression(source nodeSource, value int, scope Scope) *IntegerLiteralExpression {
	return &IntegerLiteralExpression{
		baseExpression: newBaseExpression(source, scope.GetTypeDeclaration("Int")),
		Value:          value,
	}
}

func newCharacterLiteralExpression(source nodeSource, value byte, scope Scope) *CharacterLiteralExpression {
	return &CharacterLiteralExpression{
		baseExpression: newBaseExpression(source, scope.GetTypeDeclaration("Byte")),
		Value:          value,
	}
}

func newStringLiteralExpression(source nodeSource, value string, scope Scope) *StringLiteralExpression {
	return &StringLiteralExpression{
		baseExpression: newBaseExpression(source, scope.GetTypeDeclaration("String")),
		Value:          value,
	}
}

func newBooleanLiteralExpression(source nodeSource, value bool, scope Scope) *BooleanLiteralExpression {
	return &BooleanLiteralExpression{
		baseExpression: newBaseExpression(source, scope.GetTypeDeclaration("Bool")),
		Value:          value,
	}
}

func newIdentifierExpression(source nodeSource, declaration Declaration) *IdentifierExpression {
	return &IdentifierExpression{
		baseExpression:        newBaseExpression(source),
		IdentifierDeclaration: declaration,
	}
}

func newAddExpression(source nodeSource, left Expression, right Expression) *AddExpression {
	return &AddExpression{
		dualInputExpression: dualInputExpression{
			baseExpression: newBaseExpression(source),
			Left:           left,
			Right:          right,
		},
	}
}

func newSubtractExpression(source nodeSource, left Expression, right Expression) *SubtractExpression {
	return &SubtractExpression{
		dualInputExpression: dualInputExpression{
			baseExpression: newBaseExpression(source),
			Left:           left,
			Right:          right,
		},
	}
}

func newMultiplyExpression(source nodeSource, left Expression, right Expression) *MultiplyExpression {
	return &MultiplyExpression{
		dualInputExpression: dualInputExpression{
			baseExpression: newBaseExpression(source),
			Left:           left,
			Right:          right,
		},
	}
}

func newDivideExpression(source nodeSource, left Expression, right Expression) *DivideExpression {
	return &DivideExpression{
		dualInputExpression: dualInputExpression{
			baseExpression: newBaseExpression(source),
			Left:           left,
			Right:          right,
		},
	}
}

func newEqualExpression(source nodeSource, left Expression, right Expression, scope Scope) *EqualExpression {
	return &EqualExpression{
		dualInputBoolOutputExpression: dualInputBoolOutputExpression{
			dualInputExpression: dualInputExpression{
				baseExpression: newBaseExpression(source, scope.GetTypeDeclaration("Bool")),
				Left:           left,
				Right:          right,
			},
		},
	}
}

func newNotEqualExpression(source nodeSource, left Expression, right Expression, scope Scope) *NotEqualExpression {
	return &NotEqualExpression{
		dualInputBoolOutputExpression: dualInputBoolOutputExpression{
			dualInputExpression: dualInputExpression{
				baseExpression: newBaseExpression(source, scope.GetTypeDeclaration("Bool")),
				Left:           left,
				Right:          right,
			},
		},
	}
}

func newLessExpression(source nodeSource, left Expression, right Expression, scope Scope) *LessExpression {
	return &LessExpression{
		dualInputBoolOutputExpression: dualInputBoolOutputExpression{
			dualInputExpression: dualInputExpression{
				baseExpression: newBaseExpression(source, scope.GetTypeDeclaration("Bool")),
				Left:           left,
				Right:          right,
			},
		},
	}
}

func newLessOrEqualExpression(source nodeSource, left Expression, right Expression, scope Scope) *LessOrEqualExpression {
	return &LessOrEqualExpression{
		dualInputBoolOutputExpression: dualInputBoolOutputExpression{
			dualInputExpression: dualInputExpression{
				baseExpression: newBaseExpression(source, scope.GetTypeDeclaration("Bool")),
				Left:           left,
				Right:          right,
			},
		},
	}
}

func newGreaterExpression(source nodeSource, left Expression, right Expression, scope Scope) *GreaterExpression {
	return &GreaterExpression{
		dualInputBoolOutputExpression: dualInputBoolOutputExpression{
			dualInputExpression: dualInputExpression{
				baseExpression: newBaseExpression(source, scope.GetTypeDeclaration("Bool")),
				Left:           left,
				Right:          right,
			},
		},
	}
}

func newGreaterOrEqualExpression(source nodeSource, left Expression, right Expression, scope Scope) *GreaterOrEqualExpression {
	return &GreaterOrEqualExpression{
		dualInputBoolOutputExpression: dualInputBoolOutputExpression{
			dualInputExpression: dualInputExpression{
				baseExpression: newBaseExpression(source, scope.GetTypeDeclaration("Bool")),
				Left:           left,
				Right:          right,
			},
		},
	}
}

func newAndExpression(source nodeSource, left Expression, right Expression, scope Scope) *AndExpression {
	return &AndExpression{
		dualInputBoolOutputExpression: dualInputBoolOutputExpression{
			dualInputExpression: dualInputExpression{
				baseExpression: newBaseExpression(source, scope.GetTypeDeclaration("Bool")),
				Left:           left,
				Right:          right,
			},
		},
	}
}

func newOrExpression(source nodeSource, left Expression, right Expression, scope Scope) *OrExpression {
	return &OrExpression{
		dualInputBoolOutputExpression: dualInputBoolOutputExpression{
			dualInputExpression: dualInputExpression{
				baseExpression: newBaseExpression(source, scope.GetTypeDeclaration("Bool")),
				Left:           left,
				Right:          right,
			},
		},
	}
}

func newNotExpression(source nodeSource, exp Expression, scope Scope) *NotExpression {
	return &NotExpression{
		baseExpression: newBaseExpression(source, scope.GetTypeDeclaration("Bool")),
		Expression:     exp,
	}
}

func newFunctionCallExpression(source nodeSource, callSource Expression, parameters []Expression) *FunctionCallExpression {
	return &FunctionCallExpression{
		baseExpression: newBaseExpression(source),
		CallSource:     callSource,
		Parameters:     parameters,
	}
}

func (e *IdentifierExpression) ResultingTypeDeclarations() ([]*TypeDeclaration, error) {
	if len(e.baseExpression.typeDeclarations) != 0 {
		return e.baseExpression.typeDeclarations, nil
	}

	switch d := e.IdentifierDeclaration.(type) {
	case *VariableDeclaration:
		e.baseExpression.typeDeclarations = []*TypeDeclaration{d.TypeDeclaration}
		return e.baseExpression.typeDeclarations, nil
	case *FunctionDeclaration:
		tds := make([]*TypeDeclaration, 0)
		fields := d.FunctionDefinition.FunctionType.ReturnTypes
		for _, f := range fields {
			tds = append(tds, f.VariableDeclaration.TypeDeclaration)
		}

		e.baseExpression.typeDeclarations = tds
		return tds, nil
	default:
		return nil, errors.New("compiler error: unknown declaration for identifier expression")
	}
}

func (e *NotExpression) ResultingTypeDeclarations() ([]*TypeDeclaration, error) {
	tds, err := MustSingleReturnType(e.Expression)
	if err != nil {
		return nil, err
	}

	if len(e.baseExpression.typeDeclarations) != 1 {
		return nil, errors.Errorf("compiler error: expression must have 1 return type but had %d", len(tds))
	}

	if tds[0] != e.baseExpression.typeDeclarations[0] {
		return nil, errors.Errorf("can only use 'not' operator on type Bool, type %s given on line %d column %d",
			tds[0].Type.TypeName(), e.UFSourceLine(), e.UFSourceColumn())
	}

	return tds, nil
}

func (e *FunctionCallExpression) ResultingTypeDeclarations() ([]*TypeDeclaration, error) {
	if len(e.typeDeclarations) != 0 {
		return e.typeDeclarations, nil
	}

	idExp, ok := e.CallSource.(*IdentifierExpression)
	if !ok {
		// TODO change when function-as-first-citizen calling is implemented
		return nil, errors.New("compiler error: FunctionCallExpression.CallSource is not an IdentifierExpression")
	}

	decl, ok := idExp.IdentifierDeclaration.(*FunctionDeclaration)
	if !ok {
		// TODO change when function-as-first-citizen calling is implemented
		return nil, errors.Errorf("cannot call identifier as a function on line %d column %d", e.UFSourceLine(), e.UFSourceColumn())
	}

	funcType := decl.FunctionDefinition.FunctionType
	funcParams := funcType.Parameters
	numFuncParams := len(funcParams)
	numGivenParams := len(e.Parameters)
	if numGivenParams != numFuncParams {
		return nil, errors.Errorf("number of parameters mismatch: expected %d but was given %d on line %d column %d",
			numFuncParams, numGivenParams, e.UFSourceLine(), e.UFSourceColumn())
	}

	for i, exp := range e.Parameters {
		givenTypeArr, err := MustSingleReturnType(exp)
		if err != nil {
			return nil, err
		}

		givenType := givenTypeArr[0]

		expectedType := funcParams[i].VariableDeclaration.TypeDeclaration
		if givenType != expectedType {
			return nil, errors.Errorf("parameter type mismatch: expected '%s' but was given '%d' on line %d column %d",
				expectedType.Type.TypeName(), givenType.Type.TypeName(), exp.UFSourceLine(), exp.UFSourceColumn())
		}
	}

	resultTypes := make([]*TypeDeclaration, 0)
	for _, f := range funcType.ReturnTypes {
		resultTypes = append(resultTypes, f.VariableDeclaration.TypeDeclaration)
	}

	e.typeDeclarations = resultTypes
	return resultTypes, nil
}

func (e baseExpression) ResultingTypeDeclarations() ([]*TypeDeclaration, error) {
	if len(e.typeDeclarations) == 0 {
		return nil, errors.New("compiler error: baseExpression.typeDeclarations was empty")
	}

	return e.typeDeclarations, nil
}

func (e dualInputExpression) ResultingTypeDeclarations() ([]*TypeDeclaration, error) {
	if len(e.baseExpression.typeDeclarations) != 0 {
		return e.baseExpression.typeDeclarations, nil
	}

	tds1, err := MustSingleReturnType(e.Left)
	if err != nil {
		return nil, err
	}

	tds2, err := MustSingleReturnType(e.Right)
	if err != nil {
		return nil, err
	}

	if tds1[0] != tds2[1] {
		return nil, errors.Errorf("cannot operate for different types, '%s' and '%s', on line %d column %d",
			tds1[0].Type.TypeName(), tds2[0].Type.TypeName(), e.UFSourceLine(), e.UFSourceColumn())
	}

	e.baseExpression.typeDeclarations = tds1
	return tds1, nil
}

func (e dualInputBoolOutputExpression) ResultingTypeDeclarations() ([]*TypeDeclaration, error) {
	tds, err := MustSingleReturnType(e.dualInputExpression)
	if err != nil {
		return nil, err
	}

	if len(e.baseExpression.typeDeclarations) != 1 {
		return nil, errors.Errorf("compiler error: expression must have 1 return type but had %d", len(tds))
	}

	return e.baseExpression.typeDeclarations, nil
}

func MustSingleReturnType(expression resultingTypeDeclarations) ([]*TypeDeclaration, error) {
	tds, err := expression.ResultingTypeDeclarations()
	if err != nil {
		return nil, err
	}

	if len(tds) != 1 {
		return nil, errors.Errorf("compiler error: expression must have 1 return type but had %d", len(tds))
	}

	return tds, nil
}

func (*IntegerLiteralExpression) exprNode()   {}
func (*CharacterLiteralExpression) exprNode() {}
func (*StringLiteralExpression) exprNode()    {}
func (*BooleanLiteralExpression) exprNode()   {}
func (*IdentifierExpression) exprNode()       {}
func (*AddExpression) exprNode()              {}
func (*SubtractExpression) exprNode()         {}
func (*MultiplyExpression) exprNode()         {}
func (*DivideExpression) exprNode()           {}
func (*EqualExpression) exprNode()            {}
func (*NotEqualExpression) exprNode()         {}
func (*LessExpression) exprNode()             {}
func (*LessOrEqualExpression) exprNode()      {}
func (*GreaterExpression) exprNode()          {}
func (*GreaterOrEqualExpression) exprNode()   {}
func (*AndExpression) exprNode()              {}
func (*OrExpression) exprNode()               {}
func (*NotExpression) exprNode()              {}

func (*FunctionCallExpression) exprNode() {}
func (*FunctionCallExpression) stmtNode() {}
