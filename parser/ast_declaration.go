package parser

type VariableDeclaration struct {
	nodeSource
	TypeDeclaration *TypeDeclaration
}

type TypeDeclaration struct {
	nodeSource
	Type Type
}

type FunctionDeclaration struct {
	nodeSource
	functionDefinition *FunctionDefinition
}

type UnknownDeclaration struct {
	nodeSource
	Identifier string
	Scope      Scope // Scope of the place where the identifier was used.
}

func (*VariableDeclaration) declNode() {}
func (*TypeDeclaration) declNode()     {}
func (*FunctionDeclaration) declNode() {}
func (*UnknownDeclaration) declNode()  {}

func (*VariableDeclaration) stmtNode() {}
