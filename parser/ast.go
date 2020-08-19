package parser

type Node interface {
	UFSourceLine() int
	UFSourceColumn() int
}

type Declaration interface {
	Node
	DeclarationType() string
	declNode()
}

type Statement interface {
	Node
	stmtNode()
}

type Expression interface {
	Node
	ResultingTypeDeclarations() ([]*TypeDeclaration, error)
	exprNode()
}

type Type interface {
	TypeName() string
}

type Field struct {
	Name                string
	VariableDeclaration *VariableDeclaration // variable declaration representing this field.
}

type FunctionDefinition struct {
	FunctionType FunctionType
	Statements   []Statement
}

// Internal structure to hold source code information.
type nodeSource struct {
	line   int
	column int
}

func (n nodeSource) UFSourceLine() int {
	return n.line + 1
}

func (n nodeSource) UFSourceColumn() int {
	return n.column + 1
}
