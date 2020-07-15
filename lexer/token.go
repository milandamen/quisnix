package lexer

type TokenType int

const (
	Unknown TokenType = iota

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

	// User friendly line number (starting at 1)
	UFLine() int
	// User friendly column number (starting at 1)
	UFColumn() int
}

type basicToken struct {
	tokenType TokenType
	line      int
	column    int
}

func (t basicToken) Type() TokenType {
	return t.tokenType
}

func (t basicToken) Line() int {
	return t.line
}

func (t basicToken) Column() int {
	return t.column
}

func (t basicToken) UFLine() int {
	return t.line + 1
}

func (t basicToken) UFColumn() int {
	return t.column + 1
}

type IntegerToken struct {
	basicToken
	integer int
}

func (t IntegerToken) Integer() int {
	return t.integer
}

type CharacterToken struct {
	basicToken
	character byte
}

func (t CharacterToken) Character() byte {
	return t.character
}

type StringToken struct {
	basicToken
	string string
}

func (t StringToken) String() string {
	return t.string
}

type IdentifierToken struct {
	basicToken
	identifier string
}

func (t IdentifierToken) Identifier() string {
	return t.identifier
}

func GetTokenTypeString(tt TokenType) string {
	switch tt {
	case Integer:
		return "<integer>"
	case Character:
		return "<character>"
	case String:
		return "<string>"
	case Identifier:
		return "<identifier>"
	case Add:
		return "+"
	case Subtract:
		return "-"
	case Multiply:
		return "*"
	case Divide:
		return "/"
	case Assign:
		return "="
	case AddAssign:
		return "+="
	case SubtractAssign:
		return "-="
	case Increment:
		return "++"
	case Decrement:
		return "--"
	case Equal:
		return "=="
	case NotEqual:
		return "!="
	case Less:
		return "<"
	case LessOrEqual:
		return "<="
	case Greater:
		return ">"
	case GreaterOrEqual:
		return ">="
	case And:
		return "&&"
	case Or:
		return "||"
	case Not:
		return "!"
	case LeftParenthesis:
		return "("
	case RightParenthesis:
		return ")"
	case LeftBrace:
		return "{"
	case RightBrace:
		return "}"
	case LeftBracket:
		return "["
	case RightBracket:
		return "]"
	case Comma:
		return ","
	case Period:
		return "."
	case Semicolon:
		return ";"
	case Var:
		return "var"
	case Type:
		return "type"
	case AnyType:
		return "anytype"
	case Func:
		return "func"
	case If:
		return "if"
	case Else:
		return "else"
	case Return:
		return "return"
	case For:
		return "for"
	case While:
		return "while"
	default:
		return "<unknown>"
	}
}
