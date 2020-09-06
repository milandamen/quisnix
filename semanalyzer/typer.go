package semanalyzer

import (
	"github.com/milandamen/quisnix/parser"
	"github.com/pkg/errors"
)

type Typer struct{}

func (t *Typer) Execute(declarations []parser.Declaration, scope parser.Scope) error {
	for _, decl := range declarations {
		var err error
		switch d := decl.(type) {
		case *parser.FunctionDeclaration:
			err = t.checkFunctionDeclaration(d, scope)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (t *Typer) checkFunctionDeclaration(decl *parser.FunctionDeclaration, scope parser.Scope) error {
	funcReturnTypes := make([]*parser.TypeDeclaration, 0)
	for _, f := range decl.FunctionDefinition.FunctionType.ReturnTypes {
		funcReturnTypes = append(funcReturnTypes, f.VariableDeclaration.TypeDeclaration)
	}

	err := t.checkStatements(decl.FunctionDefinition.Statements, funcReturnTypes, scope)
	if err != nil {
		return err
	}

	if len(funcReturnTypes) > 0 {
		numStmts := len(decl.FunctionDefinition.Statements)
		if numStmts != 0 {
			stmt := decl.FunctionDefinition.Statements[numStmts-1]
			if _, ok := stmt.(*parser.ReturnStatement); !ok {
				return errors.Errorf("function should return values on line %d column %d",
					decl.UFSourceLine(), decl.UFSourceColumn())
			}
		} else {
			return errors.Errorf("function should return values on line %d column %d",
				decl.UFSourceLine(), decl.UFSourceColumn())
		}
	}

	return nil
}

func (t *Typer) checkStatements(statements []parser.Statement, funcReturnTypes []*parser.TypeDeclaration, scope parser.Scope) error {
	numStmts := len(statements)
	for i, stmt := range statements {
		if sv, ok := stmt.(parser.StatementHavingVariableDeclaration); ok {
			v, ok := sv.GetVariableDeclaration().(*parser.VariableDeclaration)
			if !ok {
				return errors.New("compiler error: declaration of statement should be type VariableDeclaration")
			}

			switch s := stmt.(type) {
			case *parser.AssignStatement:
				resultTypes, err := parser.MustSingleReturnType(s.Expression)
				if err != nil {
					return err
				}

				if v.TypeDeclaration != resultTypes[0] {
					return errors.Errorf("type mismatch: expected '%s' but was given '%s' on line %d column %d",
						v.TypeDeclaration.Type.TypeName(), resultTypes[0].Type.TypeName(), s.UFSourceLine(), s.UFSourceColumn())
				}
			case *parser.AddAssignStatement:
				if v.TypeDeclaration != scope.SearchTypeDeclaration("Int") {
					return errors.Errorf("cannot add to variable with type '%s' on line %d column %d",
						v.TypeDeclaration.Type.TypeName(), s.UFSourceLine(), s.UFSourceColumn())
				}

				resultTypes, err := parser.MustSingleReturnType(s.Expression)
				if err != nil {
					return err
				}

				if v.TypeDeclaration != resultTypes[0] {
					return errors.Errorf("type mismatch: expected '%s' but was given '%s' on line %d column %d",
						v.TypeDeclaration.Type.TypeName(), resultTypes[0].Type.TypeName(), s.UFSourceLine(), s.UFSourceColumn())
				}
			case *parser.SubtractAssignStatement:
				if v.TypeDeclaration != scope.SearchTypeDeclaration("Int") {
					return errors.Errorf("cannot subtract from variable with type '%s' on line %d column %d",
						v.TypeDeclaration.Type.TypeName(), s.UFSourceLine(), s.UFSourceColumn())
				}

				resultTypes, err := parser.MustSingleReturnType(s.Expression)
				if err != nil {
					return err
				}

				if v.TypeDeclaration != resultTypes[0] {
					return errors.Errorf("type mismatch: expected '%s' but was given '%s' on line %d column %d",
						v.TypeDeclaration.Type.TypeName(), resultTypes[0].Type.TypeName(), s.UFSourceLine(), s.UFSourceColumn())
				}
			case *parser.IncrementStatement:
				if v.TypeDeclaration != scope.SearchTypeDeclaration("Int") {
					return errors.Errorf("cannot increment variable with type '%s' on line %d column %d",
						v.TypeDeclaration.Type.TypeName(), s.UFSourceLine(), s.UFSourceColumn())
				}
			case *parser.DecrementStatement:
				if v.TypeDeclaration != scope.SearchTypeDeclaration("Int") {
					return errors.Errorf("cannot decrement variable with type '%s' on line %d column %d",
						v.TypeDeclaration.Type.TypeName(), s.UFSourceLine(), s.UFSourceColumn())
				}
			}
		} else if sc, ok := stmt.(parser.StatementHavingCondition); ok {
			cond := sc.GetCondition()
			resultTypes, err := parser.MustSingleReturnType(cond)
			if err != nil {
				return err
			}

			if resultTypes[0] != scope.SearchTypeDeclaration("Bool") {
				return errors.Errorf("condition must result with type 'Bool' on line %d column %d",
					cond.UFSourceLine(), cond.UFSourceColumn())
			}

			switch s := stmt.(type) {
			case *parser.IfStatement:
				if err := t.checkStatements(s.ThenStatements, funcReturnTypes, scope); err != nil {
					return err
				}
				if err := t.checkStatements(s.ElseStatements, funcReturnTypes, scope); err != nil {
					return err
				}
			case *parser.ForStatement:
				if err := t.checkStatements(s.Statements, funcReturnTypes, scope); err != nil {
					return err
				}
			case *parser.WhileStatement:
				if err := t.checkStatements(s.Statements, funcReturnTypes, scope); err != nil {
					return err
				}
			}
		}

		if i == numStmts-1 {
			if sr, ok := stmt.(*parser.ReturnStatement); ok {
				if len(sr.ReturnExpressions) != len(funcReturnTypes) {
					return errors.Errorf("number of return types mismatch: expected %d but was given %d on line %d column %d",
						len(funcReturnTypes), len(sr.ReturnExpressions), sr.UFSourceLine(), sr.UFSourceColumn())
				}

				for i, expectedType := range funcReturnTypes {
					exp := sr.ReturnExpressions[i]
					givenTypeArr, err := parser.MustSingleReturnType(exp)
					if err != nil {
						return err
					}

					givenType := givenTypeArr[0]
					if givenType != expectedType {
						return errors.Errorf("return type mismatch: expected '%s' but was given '%d' on line %d column %d",
							expectedType.Type.TypeName(), givenType.Type.TypeName(), exp.UFSourceLine(), exp.UFSourceColumn())
					}
				}
			}
		}
	}

	return nil
}
