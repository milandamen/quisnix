package parser

type AssignStatement struct {
	nodeSource
	VariableDeclaration Declaration
	Expression          Expression
}

type AddAssignStatement struct {
	nodeSource
	VariableDeclaration Declaration
	Expression          Expression
}

type SubtractAssignStatement struct {
	nodeSource
	VariableDeclaration Declaration
	Expression          Expression
}

type IncrementStatement struct {
	nodeSource
	VariableDeclaration Declaration
}

type DecrementStatement struct {
	nodeSource
	VariableDeclaration Declaration
}

type IfStatement struct {
	nodeSource
	Condition      Expression
	ThenStatements []Statement
	ElseStatements []Statement
}

type ForStatement struct {
	nodeSource
	Init       Statement
	Condition  Expression
	LoopAction Statement
	Statements []Statement
}

type WhileStatement struct {
	nodeSource
	Condition  Expression
	Statements []Statement
}

func (*AssignStatement) stmtNode()         {}
func (*AddAssignStatement) stmtNode()      {}
func (*SubtractAssignStatement) stmtNode() {}
func (*IncrementStatement) stmtNode()      {}
func (*DecrementStatement) stmtNode()      {}
func (*IfStatement) stmtNode()             {}
func (*ForStatement) stmtNode()            {}
func (*WhileStatement) stmtNode()          {}
