package parser

type StatementHavingVariableDeclaration interface {
	GetVariableDeclaration() Declaration
	SetVariableDeclaration(declaration Declaration)
}

type StatementHavingCondition interface {
	GetCondition() Expression
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

func (s *AssignStatement) GetVariableDeclaration() Declaration {
	return s.VariableDeclaration
}

func (s *AssignStatement) SetVariableDeclaration(declaration Declaration) {
	s.VariableDeclaration = declaration
}

func (s *AddAssignStatement) GetVariableDeclaration() Declaration {
	return s.VariableDeclaration
}

func (s *AddAssignStatement) SetVariableDeclaration(declaration Declaration) {
	s.VariableDeclaration = declaration
}

func (s *SubtractAssignStatement) GetVariableDeclaration() Declaration {
	return s.VariableDeclaration
}

func (s *SubtractAssignStatement) SetVariableDeclaration(declaration Declaration) {
	s.VariableDeclaration = declaration
}

func (s *IncrementStatement) GetVariableDeclaration() Declaration {
	return s.VariableDeclaration
}

func (s *IncrementStatement) SetVariableDeclaration(declaration Declaration) {
	s.VariableDeclaration = declaration
}

func (s *DecrementStatement) GetVariableDeclaration() Declaration {
	return s.VariableDeclaration
}

func (s *DecrementStatement) SetVariableDeclaration(declaration Declaration) {
	s.VariableDeclaration = declaration
}

func (s *IfStatement) GetCondition() Expression {
	return s.Condition
}

func (s *ForStatement) GetCondition() Expression {
	return s.Condition
}

func (s *WhileStatement) GetCondition() Expression {
	return s.Condition
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
