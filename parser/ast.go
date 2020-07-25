package parser

type Node interface {
	UFSourceLine() int
	UFSourceColumn() int
}

type Declaration interface {
	Node
	declNode()
}

type Statement interface {
	Node
	stmtNode()
}

type Expression interface {
	Node
	exprNode()
}

type Type interface {
	TypeName() string
}

type Field struct {
	Name            string
	TypeDeclaration *TypeDeclaration
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
