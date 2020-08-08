package parser

type statementHavingVariableDeclaration interface {
	getVariableDeclaration() Declaration
	setVariableDeclaration(declaration Declaration)
}

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

type ReturnStatement struct {
	nodeSource
	ReturnExpressions []Expression
}

func (s *AssignStatement) getVariableDeclaration() Declaration {
	return s.VariableDeclaration
}

func (s *AssignStatement) setVariableDeclaration(declaration Declaration) {
	s.VariableDeclaration = declaration
}

func (s *AddAssignStatement) getVariableDeclaration() Declaration {
	return s.VariableDeclaration
}

func (s *AddAssignStatement) setVariableDeclaration(declaration Declaration) {
	s.VariableDeclaration = declaration
}

func (s *SubtractAssignStatement) getVariableDeclaration() Declaration {
	return s.VariableDeclaration
}

func (s *SubtractAssignStatement) setVariableDeclaration(declaration Declaration) {
	s.VariableDeclaration = declaration
}

func (s *IncrementStatement) getVariableDeclaration() Declaration {
	return s.VariableDeclaration
}

func (s *IncrementStatement) setVariableDeclaration(declaration Declaration) {
	s.VariableDeclaration = declaration
}

func (s *DecrementStatement) getVariableDeclaration() Declaration {
	return s.VariableDeclaration
}

func (s *DecrementStatement) setVariableDeclaration(declaration Declaration) {
	s.VariableDeclaration = declaration
}

func (*AssignStatement) stmtNode()         {}
func (*AddAssignStatement) stmtNode()      {}
func (*SubtractAssignStatement) stmtNode() {}
func (*IncrementStatement) stmtNode()      {}
func (*DecrementStatement) stmtNode()      {}
func (*IfStatement) stmtNode()             {}
func (*ForStatement) stmtNode()            {}
func (*WhileStatement) stmtNode()          {}
func (*ReturnStatement) stmtNode()         {}
