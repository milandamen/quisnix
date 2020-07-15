package parser

import (
	"github.com/milandamen/quisnix/lexer"
	"github.com/pkg/errors"
)

type Parser struct {
	tokens   []lexer.Token
	tokenPos int
}

func (p *Parser) Parse(tokens []lexer.Token) error {
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
		// TODO parse
		return nil, nil
	default:
		return nil, errors.Errorf("unexpected token '%s' at line %d column %d: expected: 'func'", lexer.GetTokenTypeString(tokenType), token.UFLine(), token.UFColumn())
	}
}

func (p *Parser) getNextToken() lexer.Token {
	if p.tokenPos >= len(p.tokens) {
		return nil
	}

	t := p.tokens[p.tokenPos]
	p.tokenPos++
	return t
}
