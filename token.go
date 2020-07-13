package quisnix

type TokenType int

const (
	None TokenType = iota

	// Literal
	Integer   // 12345
	Character // 'a'
	String    // "abc"

	// Identifier
	Identifier // main

	// Math
	Add      // +
	Subtract // -
	Multiply // *
	Divide   // /

	// Assignment
	Assign         // =
	AddAssign      // +=
	SubtractAssign // -=
	Increment      // ++
	Decrement      // --

	// Comparison
	Equal          // ==
	NotEqual       // !=
	Less           // <
	LessOrEqual    // <=
	Greater        // >
	GreaterOrEqual // >=
	And            // &&
	Or             // ||

	// Boolean
	Not // !

	// Delimiting
	LeftParenthesis  // (
	RightParenthesis // )
	LeftBrace        // {
	RightBrace       // }
	LeftBracket      // [
	RightBracket     // ]
	Comma            // ,
	Period           // .
	Semicolon        // ;

	// Keyword
	Var
	Type
	AnyType
	Func
	If
	Else
	Return
	For
	While
)

var (
	keywordMap = map[string]TokenType{
		"var":     Var,
		"type":    Type,
		"anytype": AnyType,
		"func":    Func,
		"if":      If,
		"else":    Else,
		"return":  Return,
		"for":     For,
		"while":   While,
	}
)

type Token interface {
	Type() TokenType

	Line() int
	Column() int
}

type BasicToken struct {
	tokenType TokenType
	line      int
	column    int
}

func (t BasicToken) Type() TokenType {
	return t.tokenType
}

func (t BasicToken) Line() int {
	return t.line
}

func (t BasicToken) Column() int {
	return t.column
}

type IntegerToken struct {
	BasicToken
	integer int
}

func (t IntegerToken) Integer() int {
	return t.integer
}

type CharacterToken struct {
	BasicToken
	character byte
}

func (t CharacterToken) Character() byte {
	return t.character
}

type StringToken struct {
	BasicToken
	string string
}

func (t StringToken) String() string {
	return t.string
}

type IdentifierToken struct {
	BasicToken
	identifier string
}

func (t IdentifierToken) Identifier() string {
	return t.identifier
}
