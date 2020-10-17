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
	FunctionDefinition *FunctionDefinition
	Name               string
}

type UnknownDeclaration struct {
	nodeSource
	Identifier string
	Scope      Scope // Scope of the place where the identifier was used.
}

func (*VariableDeclaration) DeclarationType() string {
	return "variable"
}

func (*TypeDeclaration) DeclarationType() string {
	return "type"
}

func (*FunctionDeclaration) DeclarationType() string {
	return "function"
}

func (*UnknownDeclaration) DeclarationType() string {
	return "unknown"
}

func (*VariableDeclaration) declNode() {}
func (*TypeDeclaration) declNode()     {}
func (*FunctionDeclaration) declNode() {}
func (*UnknownDeclaration) declNode()  {}

func (*VariableDeclaration) stmtNode() {}
