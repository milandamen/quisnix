package parser

type Node interface {
	SourceLine() int
	SourceColumn() int
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

type Identifier struct {
	Name string
}

type Field struct {
	Name string
	Type Type
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

func (n nodeSource) SourceLine() int {
	return n.line
}

func (n nodeSource) SourceColumn() int {
	return n.column
}
