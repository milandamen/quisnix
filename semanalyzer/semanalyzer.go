package semanalyzer

import (
	"errors"

	"github.com/milandamen/quisnix/parser"
)

type SemAnalyzer struct{}

func (s *SemAnalyzer) Analyze(declarations []parser.Declaration, scope parser.Scope) (*parser.FunctionDeclaration, error) {
	t := Typer{}
	if err := t.Execute(declarations, scope); err != nil {
		return nil, err
	}

	mainFunc, err := s.findMainFunction(scope)
	if err != nil {
		return nil, err
	}

	return mainFunc, nil
}

func (s *SemAnalyzer) findMainFunction(scope parser.Scope) (*parser.FunctionDeclaration, error) {
	decl := scope.SearchFunctionDeclaration("main")
	if decl == nil {
		return nil, errors.New("must have a 'main' function")
	}

	// TODO check arguments or return types if needed

	return decl, nil
}
