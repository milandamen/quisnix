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
		nodeSource: makeNodeSource(startToken),
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
	// TODO when to stop parsing statements, maybe peek next token first?
	for true {
		token := p.getNextToken()
		if token == nil {
			break
		}

		switch token.Type() {
		case lexer.If:
			conditionExp, err := p.parseExpression(0)
			if err != nil {
				return nil, errors.Wrap(err, "could not parse if statement condition")
			}

			lbToken := p.getNextToken()
			if lbToken == nil {
				return nil, unexpectedEOF()
			}
			if lbToken.Type() != lexer.LeftBrace {
				return nil, unexpectedTokenError(lbToken, lexer.LeftBrace)
			}

			stmts, err := p.parseStatements()
			if err != nil {
				return nil, err
			}

			statements = append(statements, IfStatement{
				nodeSource:     makeNodeSource(token),
				Condition:      conditionExp,
				ThenStatements: stmts,
				ElseStatements: make([]Statement, 0), // TODO implement
			})
		default:
			return nil, unexpectedTokenError(token, lexer.Assign, lexer.AddAssign, lexer.SubtractAssign, lexer.Increment, lexer.Decrement, lexer.If, lexer.For, lexer.While)
		}
	}

	return statements, nil
}

func (p *Parser) parseExpression(prevOperatorPrecedence int) (Expression, error) {
	token := p.getNextToken()
	if token == nil {
		return nil, unexpectedEOF()
	}

	var notExp *NotExpression
	if token.Type() == lexer.Not {
		notExp = &NotExpression{
			nodeSource: makeNodeSource(token),
			Expression: nil,
		}
	}

	var exp Expression
	switch token.Type() {
	case lexer.Integer:
		intToken, ok := token.(lexer.IntegerToken)
		if !ok {
			return nil, unexpectedTokenCastError(token)
		}

		exp = IntegerLiteralExpression{
			nodeSource: makeNodeSource(token),
			Value:      intToken.Integer(),
		}
	case lexer.Character:
		charToken, ok := token.(lexer.CharacterToken)
		if !ok {
			return nil, unexpectedTokenCastError(token)
		}

		exp = CharacterLiteralExpression{
			nodeSource: makeNodeSource(token),
			Value:      charToken.Character(),
		}
	case lexer.String:
		stringToken, ok := token.(lexer.StringToken)
		if !ok {
			return nil, unexpectedTokenCastError(token)
		}

		exp = StringLiteralExpression{
			nodeSource: makeNodeSource(token),
			Value:      stringToken.String(),
		}
	case lexer.True:
		exp = BooleanLiteralExpression{
			nodeSource: makeNodeSource(token),
			Value:      true,
		}
	case lexer.False:
		exp = BooleanLiteralExpression{
			nodeSource: makeNodeSource(token),
			Value:      false,
		}
	case lexer.Identifier:
		idToken, ok := token.(lexer.IdentifierToken)
		if !ok {
			return nil, unexpectedTokenCastError(token)
		}

		exp = IdentifierExpression{
			nodeSource: makeNodeSource(token),
			Identifier: Identifier{idToken.Identifier()},
		}
	case lexer.LeftParenthesis:
		var err error
		exp, err = p.parseParenthesizedExpression()
		if err != nil {
			return nil, err
		}
	default:
		return nil, unexpectedTokenError(token, lexer.Integer, lexer.Character, lexer.String, lexer.True, lexer.False, lexer.Identifier, lexer.LeftParenthesis)
	}

	if notExp != nil {
		notExp.Expression = exp
		exp = notExp
	}

	for true {
		pToken := p.peekNextToken()
		if pToken == nil {
			break
		}

		oToken, ok := pToken.(lexer.OperatorToken)
		if !ok {
			break
		}

		p.getNextToken()
		precedence := oToken.OperatorPrecedence()
		if precedence <= prevOperatorPrecedence {
			break
		}

		exp2, err := p.parseExpression(precedence)
		if err != nil {
			return nil, err
		}

		switch oToken.Type() {
		case lexer.Multiply:
			exp = MultiplyExpression{nodeSource: makeNodeSource(oToken), Left: exp, Right: exp2}
		case lexer.Divide:
			exp = DivideExpression{nodeSource: makeNodeSource(oToken), Left: exp, Right: exp2}
		case lexer.Add:
			exp = AddExpression{nodeSource: makeNodeSource(oToken), Left: exp, Right: exp2}
		case lexer.Subtract:
			exp = SubtractExpression{nodeSource: makeNodeSource(oToken), Left: exp, Right: exp2}
		case lexer.Equal:
			exp = EqualExpression{nodeSource: makeNodeSource(oToken), Left: exp, Right: exp2}
		case lexer.NotEqual:
			exp = NotEqualExpression{nodeSource: makeNodeSource(oToken), Left: exp, Right: exp2}
		case lexer.Less:
			exp = LessExpression{nodeSource: makeNodeSource(oToken), Left: exp, Right: exp2}
		case lexer.LessOrEqual:
			exp = LessOrEqualExpression{nodeSource: makeNodeSource(oToken), Left: exp, Right: exp2}
		case lexer.Greater:
			exp = GreaterExpression{nodeSource: makeNodeSource(oToken), Left: exp, Right: exp2}
		case lexer.GreaterOrEqual:
			exp = GreaterOrEqualExpression{nodeSource: makeNodeSource(oToken), Left: exp, Right: exp2}
		case lexer.And:
			exp = AndExpression{nodeSource: makeNodeSource(oToken), Left: exp, Right: exp2}
		case lexer.Or:
			exp = OrExpression{nodeSource: makeNodeSource(oToken), Left: exp, Right: exp2}
		default:
			return nil, unexpectedTokenError(oToken, lexer.Multiply, lexer.Divide, lexer.Add, lexer.Subtract,
				lexer.Equal, lexer.NotEqual, lexer.Less, lexer.LessOrEqual, lexer.Greater, lexer.GreaterOrEqual,
				lexer.And, lexer.Or)
		}
	}

	return exp, nil
}

func (p *Parser) parseParenthesizedExpression() (Expression, error) {
	exp, err := p.parseExpression(0)
	if err != nil {
		return nil, err
	}

	token := p.getNextToken()
	if token.Type() != lexer.RightParenthesis {
		return nil, unexpectedTokenError(token, lexer.RightParenthesis)
	}

	return exp, nil
}

func (p *Parser) getNextToken() lexer.Token {
	if p.tokenPos >= len(p.tokens) {
		return nil
	}

	t := p.tokens[p.tokenPos]
	p.tokenPos++
	return t
}

func (p *Parser) peekNextToken() lexer.Token {
	if p.tokenPos+1 >= len(p.tokens) {
		return nil
	}

	return p.tokens[p.tokenPos+1]
}

func makeNodeSource(token lexer.Token) nodeSource {
	return nodeSource{
		line:   token.Line(),
		column: token.Column(),
	}
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
