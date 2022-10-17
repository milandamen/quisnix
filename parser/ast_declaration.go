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
	MachineName        string
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

func NewFunctionDeclaration(ns nodeSource, funcDef *FunctionDefinition, name string) *FunctionDeclaration {
	return &FunctionDeclaration{
		nodeSource:         ns,
		FunctionDefinition: funcDef,
		Name:               name,

		// So it won't conflict with other functions we link into our binary. We should prefix with package path later.
		MachineName: "qx_uf_" + name,
	}
}
