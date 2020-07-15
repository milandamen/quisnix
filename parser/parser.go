package parser

import (
	"strings"

	"github.com/milandamen/quisnix/lexer"
	"github.com/pkg/errors"
)

type Parser struct {
	tokens   []lexer.Token
	tokenPos int
}

func (p *Parser) Parse(tokens []lexer.Token) error {
	p.tokens = tokens
	p.tokenPos = 0

	topLevelNodes := make([]Node, 0)

	for true {
		tln, err := p.parseTopLevel()
		if err != nil {
			return err
		}
		if tln == nil {
			break
		}

		topLevelNodes = append(topLevelNodes, tln)
	}

	// TODO do something with topLevelNodes
	return nil
}

func (p *Parser) parseTopLevel() (Node, error) {
	token := p.getNextToken()
	if token == nil {
		return nil, nil
	}

	tokenType := token.Type()
	switch tokenType {
	case lexer.Func:
		node, err := p.parseFunctionDeclaration(token)
		return node, errors.Wrapf(err, "could not parse function declaration at line %d column %d", token.UFLine(), token.UFColumn())
	default:
		return nil, unexpectedTokenError(token, lexer.Func)
	}
}

func (p *Parser) parseFunctionDeclaration(startToken lexer.Token) (Node, error) {
	token := p.getNextToken()
	if token == nil {
		return nil, unexpectedEOF()
	}
	if token.Type() != lexer.Identifier {
		return nil, unexpectedTokenError(token, lexer.Identifier)
	}

	idToken, ok := token.(lexer.IdentifierToken)
	if !ok {
		return nil, unexpectedTokenCastError(token)
	}

	def, err := p.parseFunctionDefinition()
	if err != nil {
		return nil, err
	}

	return FunctionDeclaration{
		nodeSource: nodeSource{
			line:   startToken.UFLine(),
			column: startToken.UFColumn(),
		},
		Identifier: Identifier{
			Name: idToken.Identifier(),
		},
		FunctionDefinition: def,
	}, nil
}

func (p *Parser) parseFunctionDefinition() (*FunctionDefinition, error) {
	parameters, err := p.parseFunctionParameters()
	if err != nil {
		return nil, errors.Wrap(err, "could not parse function parameters")
	}

	token := p.getNextToken()
	if token == nil {
		return nil, unexpectedEOF()
	}

	var returnTypes []Type
	if token.Type() == lexer.Identifier {
		typeToken, ok := token.(lexer.IdentifierToken)
		if !ok {
			return nil, unexpectedTokenCastError(token)
		}

		returnTypes = []Type{UnknownType{Name: typeToken.Identifier()}}
	} else if token.Type() == lexer.LeftParenthesis {
		returnTypes, err = p.parseFunctionReturnTypes()
		if err != nil {
			return nil, errors.Wrap(err, "could not parse function return types")
		}
	}

	token = p.getNextToken()
	if token == nil {
		return nil, unexpectedEOF()
	}
	if token.Type() != lexer.LeftBrace {
		return nil, unexpectedTokenError(token, lexer.LeftBrace)
	}

	statements, err := p.parseStatements()
	if err != nil {
		return nil, errors.Wrap(err, "could not parse function statements")
	}

	return &FunctionDefinition{
		FunctionType: FunctionType{
			Parameters:  parameters,
			ReturnTypes: returnTypes,
		},
		Statements: statements,
	}, nil
}

func (p *Parser) parseFunctionParameters() ([]Field, error) {
	token := p.getNextToken()
	if token == nil {
		return nil, unexpectedEOF()
	}
	if token.Type() != lexer.LeftParenthesis {
		return nil, unexpectedTokenError(token, lexer.LeftParenthesis)
	}

	parameters := make([]Field, 0)
	for true {
		token = p.getNextToken()
		if token == nil {
			return nil, unexpectedEOF()
		}

		if token.Type() == lexer.RightParenthesis {
			break
		}

		if len(parameters) > 0 {
			if token.Type() != lexer.Comma {
				return nil, unexpectedTokenError(token, lexer.RightParenthesis, lexer.Comma)
			}

			token = p.getNextToken()
			if token == nil {
				return nil, unexpectedEOF()
			}
			if token.Type() != lexer.Identifier {
				return nil, unexpectedTokenError(token, lexer.Identifier)
			}
		} else {
			if token.Type() != lexer.Identifier {
				return nil, unexpectedTokenError(token, lexer.RightParenthesis, lexer.Identifier)
			}
		}

		nameToken, ok := token.(lexer.IdentifierToken)
		if !ok {
			return nil, unexpectedTokenCastError(token)
		}

		token = p.getNextToken()
		if token == nil {
			return nil, unexpectedEOF()
		}
		if token.Type() != lexer.Identifier {
			return nil, unexpectedTokenError(token, lexer.Identifier)
		}

		typeToken, ok := token.(lexer.IdentifierToken)
		if !ok {
			return nil, unexpectedTokenCastError(token)
		}

		parameters = append(parameters, Field{
			Name: nameToken.Identifier(),
			Type: UnknownType{Name: typeToken.Identifier()},
		})
	}

	return parameters, nil
}

func (p *Parser) parseFunctionReturnTypes() ([]Type, error) {
	returnTypes := make([]Type, 0)
	for true {
		token := p.getNextToken()
		if token == nil {
			return nil, unexpectedEOF()
		}

		if len(returnTypes) > 0 {
			if token.Type() == lexer.RightParenthesis {
				break
			}

			if token.Type() != lexer.Comma {
				return nil, unexpectedTokenError(token, lexer.RightParenthesis, lexer.Comma)
			}

			token = p.getNextToken()
			if token == nil {
				return nil, unexpectedEOF()
			}
		}

		if token.Type() != lexer.Identifier {
			return nil, unexpectedTokenError(token, lexer.Identifier)
		}

		typeToken, ok := token.(lexer.IdentifierToken)
		if !ok {
			return nil, unexpectedTokenCastError(token)
		}

		returnTypes = append(returnTypes, UnknownType{Name: typeToken.Identifier()})
	}

	return returnTypes, nil
}

func (p *Parser) parseStatements() ([]Statement, error) {
	statements := make([]Statement, 0)
	for true {
		token := p.getNextToken()
		if token == nil {
			break
		}

		switch token.Type() {

		default:
			return nil, unexpectedTokenError(token, lexer.Assign, lexer.AddAssign, lexer.SubtractAssign, lexer.Increment, lexer.Decrement, lexer.If, lexer.For, lexer.While)
		}
	}

	return statements, nil
}

func (p *Parser) getNextToken() lexer.Token {
	if p.tokenPos >= len(p.tokens) {
		return nil
	}

	t := p.tokens[p.tokenPos]
	p.tokenPos++
	return t
}

func unexpectedTokenError(token lexer.Token, expectedTokenTypes ...lexer.TokenType) error {
	expectedTypes := make([]string, 0)
	for _, tt := range expectedTokenTypes {
		expectedTypes = append(expectedTypes, "'"+lexer.GetTokenTypeString(tt)+"'")
	}

	return errors.Errorf("unexpected token '%s' at line %d column %d: expected: %s",
		lexer.GetTokenTypeString(token.Type()),
		token.UFLine(),
		token.UFColumn(),
		strings.Join(expectedTypes, ", "))
}

func unexpectedTokenCastError(token lexer.Token) error {
	return errors.Errorf("could not cast token with type '%s' at line %d column %d to internal Token implementation.",
		lexer.GetTokenTypeString(token.Type()),
		token.UFLine(),
		token.UFColumn())
}

func unexpectedEOF() error {
	return errors.New("unexpected end of file")
}
