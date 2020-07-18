package parser

type VariableDeclaration struct {
	nodeSource
	Type Type
}

type TypeDeclaration struct {
	nodeSource
	Type Type
}

type FunctionDeclaration struct {
	nodeSource
	functionDefinition *FunctionDefinition
}

func (VariableDeclaration) declNode() {}
func (TypeDeclaration) declNode()     {}
func (FunctionDeclaration) declNode() {}
