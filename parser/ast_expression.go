package parser

type IntegerLiteralExpression struct {
	nodeSource
	Value int
}

type CharacterLiteralExpression struct {
	nodeSource
	Value byte
}

type StringLiteralExpression struct {
	nodeSource
	Value string
}

type AddExpression struct {
	nodeSource
	Left  Expression
	Right Expression
}

type SubtractExpression struct {
	nodeSource
	Left  Expression
	Right Expression
}

type MultiplyExpression struct {
	nodeSource
	Left  Expression
	Right Expression
}

type DivideExpression struct {
	nodeSource
	Left  Expression
	Right Expression
}

type EqualExpression struct {
	nodeSource
	Left  Expression
	Right Expression
}

type NotEqualExpression struct {
	nodeSource
	Left  Expression
	Right Expression
}

type LessExpression struct {
	nodeSource
	Left  Expression
	Right Expression
}

type LessOrEqualExpression struct {
	nodeSource
	Left  Expression
	Right Expression
}

type GreaterExpression struct {
	nodeSource
	Left  Expression
	Right Expression
}

type GreaterOrEqualExpression struct {
	nodeSource
	Left  Expression
	Right Expression
}

type AndExpression struct {
	nodeSource
	Left  Expression
	Right Expression
}

type OrExpression struct {
	nodeSource
	Left  Expression
	Right Expression
}

type NotExpression struct {
	nodeSource
	Expression Expression
}

func (IntegerLiteralExpression) exprNode()   {}
func (CharacterLiteralExpression) exprNode() {}
func (StringLiteralExpression) exprNode()    {}
func (AddExpression) exprNode()              {}
func (SubtractExpression) exprNode()         {}
func (MultiplyExpression) exprNode()         {}
func (DivideExpression) exprNode()           {}
func (EqualExpression) exprNode()            {}
func (NotEqualExpression) exprNode()         {}
func (LessExpression) exprNode()             {}
func (LessOrEqualExpression) exprNode()      {}
func (GreaterExpression) exprNode()          {}
func (GreaterOrEqualExpression) exprNode()   {}
func (AndExpression) exprNode()              {}
func (OrExpression) exprNode()               {}
func (NotExpression) exprNode()              {}