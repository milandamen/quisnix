package typer

import (
	"github.com/milandamen/quisnix/parser"
	"github.com/pkg/errors"
)

type Typer struct {
}

func (t *Typer) Execute(declarations []parser.Declaration) error {
	for _, decl := range declarations {
		var err error
		switch d := decl.(type) {
		case *parser.FunctionDeclaration:
			err = t.checkFunctionDeclaration(d)
		}

		if err != nil {
			return err
		}
	}

	// TODO go through all expressions and resolve their types
	// TODO go through all statements and expressions and check that types are comparable and assignable

	return nil
}

func (t *Typer) checkFunctionDeclaration(decl *parser.FunctionDeclaration) error {
	for _, stmt := range decl.FunctionDefinition.Statements {
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
				resultTypes, err := parser.MustSingleReturnType(s.Expression)
				if err != nil {
					return err
				}

				if v.TypeDeclaration != resultTypes[0] {
					return errors.Errorf("type mismatch: expected '%s' but was given '%s' on line %d column %d",
						v.TypeDeclaration.Type.TypeName(), resultTypes[0].Type.TypeName(), s.UFSourceLine(), s.UFSourceColumn())
				}
			case *parser.SubtractAssignStatement:
				resultTypes, err := parser.MustSingleReturnType(s.Expression)
				if err != nil {
					return err
				}

				if v.TypeDeclaration != resultTypes[0] {
					return errors.Errorf("type mismatch: expected '%s' but was given '%s' on line %d column %d",
						v.TypeDeclaration.Type.TypeName(), resultTypes[0].Type.TypeName(), s.UFSourceLine(), s.UFSourceColumn())
				}
			}
		}
	}

	return nil
}
