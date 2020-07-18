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

	// TODO add declaration to current scope
	return FunctionDeclaration{
		nodeSource: makeNodeSource(startToken),
		Identifier: Identifier{
			Name: idToken.Identifier(),
		},
		functionDefinition: def,
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

		// TODO what if its a complex type?
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

		// TODO what if its a complex type?
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
		// TODO implement For
		// TODO implement Var (and add var declaration to current scope)
		case lexer.If:
			stmt, err := p.parseIfStatement(token)
			if err != nil {
				return nil, err
			}

			statements = append(statements, stmt)
		case lexer.While:
			stmt, err := p.parseWhileStatement(token)
			if err != nil {
				return nil, err
			}

			statements = append(statements, stmt)
		case lexer.Identifier:
			idToken, ok := token.(lexer.IdentifierToken)
			if !ok {
				return nil, unexpectedTokenCastError(token)
			}

			stmt, err := p.parseIdentifierStatement(idToken)
			if err != nil {
				return nil, err
			}

			statements = append(statements, stmt)
		case lexer.RightBrace:
			break
		default:
			return nil, unexpectedTokenError(token, lexer.Identifier, lexer.If, lexer.For, lexer.While, lexer.RightBrace)
		}
	}

	return statements, nil
}

func (p *Parser) parseIfStatement(startToken lexer.Token) (Statement, error) {
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

	elseStmts := make([]Statement, 0)
	pToken := p.peekNextToken()
	if pToken == nil {
		return nil, unexpectedEOF()
	}
	if pToken.Type() == lexer.Else {
		p.getNextToken()
		lbToken = p.getNextToken()
		if lbToken == nil {
			return nil, unexpectedEOF()
		}
		if lbToken.Type() != lexer.LeftBrace {
			return nil, unexpectedTokenError(lbToken, lexer.LeftBrace) // FIXME: Currently does not support "else if"
		}

		elseStmts, err = p.parseStatements()
		if err != nil {
			return nil, err
		}
	}

	return IfStatement{
		nodeSource:     makeNodeSource(startToken),
		Condition:      conditionExp,
		ThenStatements: stmts,
		ElseStatements: elseStmts,
	}, nil
}

func (p *Parser) parseWhileStatement(startToken lexer.Token) (Statement, error) {
	conditionExp, err := p.parseExpression(0)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse while statement condition")
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

	return WhileStatement{
		nodeSource: makeNodeSource(startToken),
		Condition:  conditionExp,
		Statements: stmts,
	}, nil
}

func (p *Parser) parseIdentifierStatement(idToken lexer.IdentifierToken) (Statement, error) {
	token := p.getNextToken()
	if token == nil {
		return nil, unexpectedEOF()
	}

	var stmt Statement
	switch token.Type() {
	case lexer.Assign:
		exp, err := p.parseExpression(0)
		if err != nil {
			return nil, err
		}

		stmt = AssignStatement{
			nodeSource: makeNodeSource(idToken),
			Identifier: Identifier{idToken.Identifier()},
			Expression: exp,
		}
	case lexer.AddAssign:
		exp, err := p.parseExpression(0)
		if err != nil {
			return nil, err
		}

		stmt = AddAssignStatement{
			nodeSource: makeNodeSource(idToken),
			Identifier: Identifier{idToken.Identifier()},
			Expression: exp,
		}
	case lexer.SubtractAssign:
		exp, err := p.parseExpression(0)
		if err != nil {
			return nil, err
		}

		stmt = SubtractAssignStatement{
			nodeSource: makeNodeSource(idToken),
			Identifier: Identifier{idToken.Identifier()},
			Expression: exp,
		}
	case lexer.Increment:
		stmt = IncrementStatement{
			nodeSource: makeNodeSource(idToken),
			Identifier: Identifier{idToken.Identifier()},
		}
	case lexer.Decrement:
		stmt = DecrementStatement{
			nodeSource: makeNodeSource(idToken),
			Identifier: Identifier{idToken.Identifier()},
		}
	case lexer.LeftParenthesis:
		var err error
		stmt, err = p.parseFunctionCallExpression(idToken, IdentifierExpression{
			nodeSource: makeNodeSource(idToken),
			Identifier: Identifier{idToken.Identifier()},
		})
		if err != nil {
			return nil, err
		}
	default:
		return nil, unexpectedTokenError(token, lexer.Assign, lexer.AddAssign, lexer.SubtractAssign, lexer.Increment, lexer.Decrement, lexer.LeftParenthesis)
	}

	token = p.getNextToken()
	if token == nil {
		return nil, unexpectedEOF()
	}
	if token.Type() != lexer.Semicolon {
		return nil, unexpectedTokenError(token, lexer.Semicolon)
	}

	return stmt, nil
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

	isCallableExp := func(e Expression) bool {
		if _, ok := e.(IdentifierExpression); ok {
			return true
		}
		if _, ok := e.(FunctionCallExpression); ok {
			return true
		}
		return false
	}

	for true {
		pToken := p.peekNextToken()
		if pToken == nil {
			break
		}

		oToken, ok := pToken.(lexer.OperatorToken)
		if ok {
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
		} else {
			if pToken.Type() == lexer.LeftParenthesis && isCallableExp(exp) {
				p.getNextToken()
				var err error
				exp, err = p.parseFunctionCallExpression(pToken, exp)
				if err != nil {
					return nil, err
				}
			} else {
				break
			}
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

func (p *Parser) parseFunctionCallExpression(startToken lexer.Token, callSource Expression) (*FunctionCallExpression, error) {
	parameters := make([]Expression, 0)
	for true {
		token := p.getNextToken()
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

		exp, err := p.parseExpression(0)
		if err != nil {
			return nil, err
		}

		parameters = append(parameters, exp)
	}

	return &FunctionCallExpression{
		nodeSource: makeNodeSource(startToken),
		CallSource: callSource,
		Parameters: parameters,
	}, nil
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
