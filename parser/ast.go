package parser

type Node interface {
	SourceLine() int
	SourceColumn() int
}

type Declaration interface {
	declNode()
}

type Statement interface {
	stmtNode()
}

type Expression interface {
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
