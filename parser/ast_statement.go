package parser

type AssignStatement struct {
	nodeSource
	Identifier Identifier
	Expression Expression
}

type AddAssignStatement struct {
	nodeSource
	Identifier Identifier
	Expression Expression
}

type SubtractAssignStatement struct {
	nodeSource
	Identifier Identifier
	Expression Expression
}

type IncrementStatement struct {
	nodeSource
	Identifier Identifier
}

type DecrementStatement struct {
	nodeSource
	Identifier Identifier
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

func (AssignStatement) stmtNode()         {}
func (AddAssignStatement) stmtNode()      {}
func (SubtractAssignStatement) stmtNode() {}
func (IncrementStatement) stmtNode()      {}
func (DecrementStatement) stmtNode()      {}
func (IfStatement) stmtNode()             {}
func (ForStatement) stmtNode()            {}
func (WhileStatement) stmtNode()          {}
