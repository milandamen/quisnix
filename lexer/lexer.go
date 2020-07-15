package lexer

import (
	"bufio"
	"io"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type Lexer struct {
}

func (l Lexer) Parse(r io.Reader) ([]Token, error) {
	tokens := make([]Token, 0)
	scanner := bufio.NewScanner(r)

	lineIdx := 0
	for scanner.Scan() {
		line := scanner.Bytes()

		column := 0
		lineLen := len(line)
		for column < lineLen {
			if l.isWhitespaceCharacter(line[column]) {
				column++
				continue
			}
			if t := l.getDoubleCharacterToken(line, lineIdx, column); t != nil {
				tokens = append(tokens, t)
				column += 2
				continue
			}
			if t := l.getSingleCharacterToken(line, lineIdx, column); t != nil {
				tokens = append(tokens, t)
				column++
				continue
			}

			t, newColumn, err := l.getLiteralToken(line, lineIdx, column)
			if err != nil {
				return nil, errors.Wrapf(err, "error at line %d column %d", lineIdx+1, column+1)
			}
			if t != nil {
				tokens = append(tokens, t)
				column = newColumn
				continue
			}

			t, newColumn = l.getKeywordToken(line, lineIdx, column)
			if t != nil {
				tokens = append(tokens, t)
				column = newColumn
				continue
			}

			t, newColumn = l.getIdentifierToken(line, lineIdx, column)
			if t != nil {
				tokens = append(tokens, t)
				column = newColumn
				continue
			}

			return nil, errors.Errorf("unknown token at line %d column %d", lineIdx+1, column+1)
		}

		lineIdx++
	}

	if err := scanner.Err(); err != nil {
		return nil, errors.Wrap(err, "error scanning lines")
	}

	return tokens, nil
}

func (Lexer) isWhitespaceCharacter(char byte) bool {
	return char == ' ' || char == '\t' || char == '\r'
}

func (Lexer) getDoubleCharacterToken(line []byte, lineIdx, column int) Token {
	if column+1 >= len(line) {
		return nil
	}
	c0 := line[column]
	c1 := line[column+1]

	if c1 == '=' {
		if c0 == '+' {
			return basicToken{tokenType: AddAssign, line: lineIdx, column: column}
		}
		if c0 == '-' {
			return basicToken{tokenType: SubtractAssign, line: lineIdx, column: column}
		}
		if c0 == '=' {
			return basicToken{tokenType: Equal, line: lineIdx, column: column}
		}
		if c0 == '!' {
			return basicToken{tokenType: NotEqual, line: lineIdx, column: column}
		}
		if c0 == '<' {
			return basicToken{tokenType: LessOrEqual, line: lineIdx, column: column}
		}
		if c0 == '>' {
			return basicToken{tokenType: GreaterOrEqual, line: lineIdx, column: column}
		}
		return nil
	}

	if c0 == '+' && c1 == '+' {
		return basicToken{tokenType: Increment, line: lineIdx, column: column}
	}
	if c0 == '-' && c1 == '-' {
		return basicToken{tokenType: Decrement, line: lineIdx, column: column}
	}
	if c0 == '&' && c1 == '&' {
		return basicToken{tokenType: And, line: lineIdx, column: column}
	}
	if c0 == '|' && c1 == '|' {
		return basicToken{tokenType: Or, line: lineIdx, column: column}
	}

	return nil
}

func (Lexer) getSingleCharacterToken(line []byte, lineIdx, column int) Token {
	c0 := line[column]

	if c0 == '+' {
		return basicToken{tokenType: Add, line: lineIdx, column: column}
	}
	if c0 == '-' {
		return basicToken{tokenType: Subtract, line: lineIdx, column: column}
	}
	if c0 == '*' {
		return basicToken{tokenType: Multiply, line: lineIdx, column: column}
	}
	if c0 == '/' {
		return basicToken{tokenType: Divide, line: lineIdx, column: column}
	}
	if c0 == '=' {
		return basicToken{tokenType: Assign, line: lineIdx, column: column}
	}
	if c0 == '<' {
		return basicToken{tokenType: Less, line: lineIdx, column: column}
	}
	if c0 == '>' {
		return basicToken{tokenType: Greater, line: lineIdx, column: column}
	}
	if c0 == '!' {
		return basicToken{tokenType: Not, line: lineIdx, column: column}
	}
	if c0 == '(' {
		return basicToken{tokenType: LeftParenthesis, line: lineIdx, column: column}
	}
	if c0 == ')' {
		return basicToken{tokenType: RightParenthesis, line: lineIdx, column: column}
	}
	if c0 == '{' {
		return basicToken{tokenType: LeftBrace, line: lineIdx, column: column}
	}
	if c0 == '}' {
		return basicToken{tokenType: RightBrace, line: lineIdx, column: column}
	}
	if c0 == '[' {
		return basicToken{tokenType: LeftBracket, line: lineIdx, column: column}
	}
	if c0 == ']' {
		return basicToken{tokenType: RightBracket, line: lineIdx, column: column}
	}
	if c0 == ',' {
		return basicToken{tokenType: Comma, line: lineIdx, column: column}
	}
	if c0 == '.' {
		return basicToken{tokenType: Period, line: lineIdx, column: column}
	}
	if c0 == ';' {
		return basicToken{tokenType: Semicolon, line: lineIdx, column: column}
	}

	return nil
}

func (Lexer) getLiteralToken(line []byte, lineIdx, column int) (Token, int, error) {
	lineLen := len(line)
	c0 := line[column]
	if c0 >= '0' && c0 <= '9' {
		untilCol := column + 1
		for untilCol < lineLen {
			cx := line[untilCol]
			if cx >= '0' && cx <= '9' {
				untilCol++
			} else {
				break
			}
		}

		s := line[column:untilCol] // From column until (exclusive) untilCol
		integer, err := strconv.Atoi(string(s))
		if err != nil {
			return nil, 0, errors.Wrapf(err, "could not parse '%s' into integer", s)
		}

		return IntegerToken{
			basicToken: basicToken{
				tokenType: Integer,
				line:      lineIdx,
				column:    column,
			},
			integer: integer,
		}, untilCol, nil
	}

	if c0 == '\'' {
		if column+1 < lineLen && line[column+1] == '\'' {
			return nil, 0, errors.New("character literal can not be empty")
		}
		if column+2 >= lineLen {
			return nil, 0, errors.New("unexpected end of line before end of character literal")
		}
		if line[column+2] != '\'' {
			return nil, 0, errors.New("character literal may only be 1 character long")
		}

		return CharacterToken{
			basicToken: basicToken{
				tokenType: Character,
				line:      lineIdx,
				column:    column,
			},
			character: line[column+1],
		}, column + 3, nil
	}

	if c0 == '"' {
		untilCol := column + 1
		closed := false
		for untilCol < lineLen {
			if line[untilCol] == '"' {
				closed = true
				untilCol++
				break
			}

			untilCol++
		}

		if !closed {
			return nil, 0, errors.New("unexpected end of line before end of string literal")
		}

		s := line[column+1 : untilCol-1]
		return StringToken{
			basicToken: basicToken{
				tokenType: String,
				line:      lineIdx,
				column:    column,
			},
			string: string(s),
		}, untilCol, nil
	}

	return nil, 0, nil
}

func (Lexer) getKeywordToken(line []byte, lineIdx, column int) (Token, int) {
	lineLen := len(line)
	remainingLineString := string(line[column:])
	for keyword, tokenType := range keywordMap {
		untilCol := column + len(keyword)
		if strings.HasPrefix(remainingLineString, keyword) && (untilCol >= lineLen || !(line[untilCol] >= 'a' && line[untilCol] <= 'z')) {
			return basicToken{
				tokenType: tokenType,
				line:      lineIdx,
				column:    column,
			}, untilCol
		}
	}

	return nil, 0
}

func (Lexer) getIdentifierToken(line []byte, lineIdx, column int) (Token, int) {
	lineLen := len(line)
	untilCol := column
	for untilCol < lineLen {
		c := line[untilCol]
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || c == '_' ||
			(untilCol != column && c >= '0' && c <= '9') {

			untilCol++
		} else {
			break
		}
	}

	if untilCol != column {
		s := line[column:untilCol]
		return IdentifierToken{
			basicToken: basicToken{
				tokenType: Identifier,
				line:      lineIdx,
				column:    column,
			},
			identifier: string(s),
		}, untilCol
	}

	return nil, 0
}
