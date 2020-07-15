package parser

type VariableDeclaration struct {
	nodeSource
	Identifier Identifier
	Type       Type
}

type TypeDeclaration struct {
	nodeSource
	Identifier Identifier
	Type       Type
}

type FunctionDeclaration struct {
	nodeSource
	Identifier         Identifier
	FunctionDefinition *FunctionDefinition
}

func (VariableDeclaration) declNode() {}
func (TypeDeclaration) declNode()     {}
func (FunctionDeclaration) declNode() {}
