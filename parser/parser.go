package parser

import (
	"strings"

	"github.com/milandamen/quisnix/lexer"
	"github.com/pkg/errors"
)

type Parser struct {
	tokens   []lexer.Token
	tokenPos int

	unknownFieldTypes           []*Field
	unknownIdentifiers          []*IdentifierExpression
	unknownIdentifierStatements []Statement
	// TODO keep UnknownTypes and resolve them once parsing the file is done
}

func (p *Parser) Parse(tokens []lexer.Token) error {
	p.tokens = tokens
	p.tokenPos = 0

	topLevelNodes := make([]Node, 0)
	builtInScope := NewBuiltInScope()
	fileScope := NewFileScope(builtInScope)

	for true {
		tln, err := p.parseTopLevel(fileScope)
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

func (p *Parser) parseTopLevel(currentScope *FileScope) (Node, error) {
	token := p.getNextToken()
	if token == nil {
		return nil, nil
	}

	tokenType := token.Type()
	switch tokenType {
	case lexer.Func:
		node, err := p.parseFunctionDeclaration(token, currentScope)
		return node, errors.Wrapf(err, "could not parse function declaration at line %d column %d", token.UFLine(), token.UFColumn())
	default:
		return nil, unexpectedTokenError(token, lexer.Func)
	}
}

func (p *Parser) parseFunctionDeclaration(startToken lexer.Token, currentScope Scope) (Node, error) {
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

	def, err := p.parseFunctionDefinition(currentScope)
	if err != nil {
		return nil, err
	}

	// TODO make sure this declaration does not clash with other identifiers in this scope
	decl := FunctionDeclaration{
		nodeSource:         makeNodeSource(startToken),
		functionDefinition: def,
	}
	currentScope.DeclareFunction(idToken.Identifier(), decl)
	return decl, nil
}

func (p *Parser) parseFunctionDefinition(currentScope Scope) (*FunctionDefinition, error) {
	parameters, err := p.parseFunctionParameters(currentScope)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse function parameters")
	}

	token := p.peekNextToken()
	if token == nil {
		return nil, unexpectedEOF()
	}

	returnTypes, err := p.parseFunctionReturnTypes(currentScope)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse function return types")
	}

	token = p.getNextToken()
	if token == nil {
		return nil, unexpectedEOF()
	}
	if token.Type() != lexer.LeftBrace {
		return nil, unexpectedTokenError(token, lexer.LeftBrace)
	}

	stmtsScope := NewBasicScope(currentScope, FunctionScopeType)
	statements, err := p.parseStatements(stmtsScope)
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

func (p *Parser) parseFunctionParameters(currentScope Scope) ([]*Field, error) {
	token := p.getNextToken()
	if token == nil {
		return nil, unexpectedEOF()
	}
	if token.Type() != lexer.LeftParenthesis {
		return nil, unexpectedTokenError(token, lexer.LeftParenthesis)
	}

	parameters := make([]*Field, 0)
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

		parameters = append(parameters, p.getTypedField(nameToken.Identifier(), typeToken.Identifier(), currentScope))
	}

	return parameters, nil
}

func (p *Parser) parseFunctionReturnTypes(currentScope Scope) ([]*Field, error) {
	token := p.peekNextToken()
	if token == nil {
		return nil, unexpectedEOF()
	}

	returnTypes := make([]*Field, 0)
	if token.Type() == lexer.Identifier {
		p.getNextToken()
		typeToken, ok := token.(lexer.IdentifierToken)
		if !ok {
			return nil, unexpectedTokenCastError(token)
		}

		return []*Field{
			p.getTypedField("", typeToken.Identifier(), currentScope),
		}, nil
	} else if token.Type() == lexer.LeftParenthesis {
		p.getNextToken()
	} else {
		return make([]*Field, 0), nil
	}

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

		returnTypes = append(returnTypes, p.getTypedField("", typeToken.Identifier(), currentScope))
	}

	return returnTypes, nil
}

func (p *Parser) getTypedField(fieldName string, typeId string, currentScope Scope) *Field {
	// TODO what if its a complex type?
	var f *Field
	decl := currentScope.SearchTypeDeclaration(typeId)
	if decl != nil {
		f = &Field{
			Name:            fieldName,
			TypeDeclaration: decl,
		}
	} else {
		// TODO make sure this declaration does not clash with others in this scope (like func or variable decl)
		decl = &TypeDeclaration{
			nodeSource: nodeSource{},
			Type:       UnknownType{Name: typeId, Scope: currentScope},
		}

		f = &Field{
			Name:            fieldName,
			TypeDeclaration: decl,
		}

		p.unknownFieldTypes = append(p.unknownFieldTypes, f)
	}

	return f
}

func (p *Parser) parseStatements(currentScope Scope) ([]Statement, error) {
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
			stmt, err := p.parseIfStatement(token, currentScope)
			if err != nil {
				return nil, err
			}

			statements = append(statements, stmt)
		case lexer.While:
			stmt, err := p.parseWhileStatement(token, currentScope)
			if err != nil {
				return nil, err
			}

			statements = append(statements, stmt)
		case lexer.Var:
			stmt, err := p.parseVariableDeclarationStatement(currentScope)
			if err != nil {
				return nil, err
			}

			statements = append(statements, stmt)
		case lexer.Identifier:
			idToken, ok := token.(lexer.IdentifierToken)
			if !ok {
				return nil, unexpectedTokenCastError(token)
			}

			stmt, err := p.parseIdentifierStatement(idToken, currentScope)
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

func (p *Parser) parseIfStatement(startToken lexer.Token, currentScope Scope) (Statement, error) {
	conditionExp, err := p.parseExpression(0, currentScope)
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

	stmtsScope := NewBasicScope(currentScope, BlockScopeType)
	stmts, err := p.parseStatements(stmtsScope)
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

		elseStmtsScope := NewBasicScope(currentScope, BlockScopeType)
		elseStmts, err = p.parseStatements(elseStmtsScope)
		if err != nil {
			return nil, err
		}
	}

	return &IfStatement{
		nodeSource:     makeNodeSource(startToken),
		Condition:      conditionExp,
		ThenStatements: stmts,
		ElseStatements: elseStmts,
	}, nil
}

func (p *Parser) parseWhileStatement(startToken lexer.Token, currentScope Scope) (Statement, error) {
	conditionExp, err := p.parseExpression(0, currentScope)
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

	stmtsScope := NewBasicScope(currentScope, BlockScopeType)
	stmts, err := p.parseStatements(stmtsScope)
	if err != nil {
		return nil, err
	}

	return &WhileStatement{
		nodeSource: makeNodeSource(startToken),
		Condition:  conditionExp,
		Statements: stmts,
	}, nil
}

func (p *Parser) parseVariableDeclarationStatement(currentScope Scope) (Statement, error) {
	token := p.getNextToken()
	if token == nil {
		return nil, unexpectedEOF()
	}

	// TODO make sure this declaration does not clash with others in this scope
}

func (p *Parser) parseIdentifierStatement(idToken lexer.IdentifierToken, currentScope Scope) (Statement, error) {
	token := p.getNextToken()
	if token == nil {
		return nil, unexpectedEOF()
	}

	id := idToken.Identifier()
	var decl Declaration = currentScope.SearchVariableDeclaration(id)
	var addUnknownIdentifierStmt bool
	if decl == nil {
		addUnknownIdentifierStmt = true
		decl = &UnknownDeclaration{
			nodeSource: makeNodeSource(idToken),
			Identifier: id,
			Scope:      currentScope,
		}
	}

	var stmt Statement
	switch token.Type() {
	case lexer.Assign:
		exp, err := p.parseExpression(0, currentScope)
		if err != nil {
			return nil, err
		}

		stmt = &AssignStatement{
			nodeSource:          makeNodeSource(idToken),
			VariableDeclaration: decl,
			Expression:          exp,
		}
	case lexer.AddAssign:
		exp, err := p.parseExpression(0, currentScope)
		if err != nil {
			return nil, err
		}

		stmt = &AddAssignStatement{
			nodeSource:          makeNodeSource(idToken),
			VariableDeclaration: decl,
			Expression:          exp,
		}
	case lexer.SubtractAssign:
		exp, err := p.parseExpression(0, currentScope)
		if err != nil {
			return nil, err
		}

		stmt = &SubtractAssignStatement{
			nodeSource:          makeNodeSource(idToken),
			VariableDeclaration: decl,
			Expression:          exp,
		}
	case lexer.Increment:
		stmt = &IncrementStatement{
			nodeSource:          makeNodeSource(idToken),
			VariableDeclaration: decl,
		}
	case lexer.Decrement:
		stmt = &DecrementStatement{
			nodeSource:          makeNodeSource(idToken),
			VariableDeclaration: decl,
		}
	case lexer.LeftParenthesis:
		var addUnknownIdentifierExp bool
		if addUnknownIdentifierStmt {
			addUnknownIdentifierStmt = false
			funcDecl := currentScope.SearchFunctionDeclaration(id)
			if funcDecl != nil {
				decl = funcDecl
			} else {
				addUnknownIdentifierExp = true
			}
		}

		exp := &IdentifierExpression{
			nodeSource:            makeNodeSource(idToken),
			IdentifierDeclaration: decl,
		}

		if addUnknownIdentifierExp {
			p.unknownIdentifiers = append(p.unknownIdentifiers, exp)
		}

		var err error
		stmt, err = p.parseFunctionCallExpression(idToken, exp, currentScope)
		if err != nil {
			return nil, err
		}
	default:
		return nil, unexpectedTokenError(token, lexer.Assign, lexer.AddAssign, lexer.SubtractAssign, lexer.Increment, lexer.Decrement, lexer.LeftParenthesis)
	}

	if addUnknownIdentifierStmt {
		p.unknownIdentifierStatements = append(p.unknownIdentifierStatements, stmt)
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

func (p *Parser) parseExpression(prevOperatorPrecedence int, currentScope Scope) (Expression, error) {
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

		exp = &IntegerLiteralExpression{
			nodeSource: makeNodeSource(token),
			Value:      intToken.Integer(),
		}
	case lexer.Character:
		charToken, ok := token.(lexer.CharacterToken)
		if !ok {
			return nil, unexpectedTokenCastError(token)
		}

		exp = &CharacterLiteralExpression{
			nodeSource: makeNodeSource(token),
			Value:      charToken.Character(),
		}
	case lexer.String:
		stringToken, ok := token.(lexer.StringToken)
		if !ok {
			return nil, unexpectedTokenCastError(token)
		}

		exp = &StringLiteralExpression{
			nodeSource: makeNodeSource(token),
			Value:      stringToken.String(),
		}
	case lexer.True:
		exp = &BooleanLiteralExpression{
			nodeSource: makeNodeSource(token),
			Value:      true,
		}
	case lexer.False:
		exp = &BooleanLiteralExpression{
			nodeSource: makeNodeSource(token),
			Value:      false,
		}
	case lexer.Identifier:
		idToken, ok := token.(lexer.IdentifierToken)
		if !ok {
			return nil, unexpectedTokenCastError(token)
		}

		id := idToken.Identifier()
		var decl Declaration = currentScope.SearchVariableDeclaration(id)
		if decl == nil {
			decl = currentScope.SearchFunctionDeclaration(id)
		}
		if decl == nil {
			decl = &UnknownDeclaration{
				nodeSource: makeNodeSource(idToken),
				Identifier: id,
				Scope:      currentScope,
			}
			idExp := &IdentifierExpression{
				nodeSource:            makeNodeSource(token),
				IdentifierDeclaration: decl,
			}
			p.unknownIdentifiers = append(p.unknownIdentifiers, idExp)
			exp = idExp
		} else {
			exp = &IdentifierExpression{
				nodeSource:            makeNodeSource(token),
				IdentifierDeclaration: decl,
			}
		}
	case lexer.LeftParenthesis:
		var err error
		exp, err = p.parseParenthesizedExpression(currentScope)
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
		if _, ok := e.(*IdentifierExpression); ok {
			return true
		}
		if _, ok := e.(*FunctionCallExpression); ok {
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

			exp2, err := p.parseExpression(precedence, currentScope)
			if err != nil {
				return nil, err
			}

			switch oToken.Type() {
			case lexer.Multiply:
				exp = &MultiplyExpression{nodeSource: makeNodeSource(oToken), Left: exp, Right: exp2}
			case lexer.Divide:
				exp = &DivideExpression{nodeSource: makeNodeSource(oToken), Left: exp, Right: exp2}
			case lexer.Add:
				exp = &AddExpression{nodeSource: makeNodeSource(oToken), Left: exp, Right: exp2}
			case lexer.Subtract:
				exp = &SubtractExpression{nodeSource: makeNodeSource(oToken), Left: exp, Right: exp2}
			case lexer.Equal:
				exp = &EqualExpression{nodeSource: makeNodeSource(oToken), Left: exp, Right: exp2}
			case lexer.NotEqual:
				exp = &NotEqualExpression{nodeSource: makeNodeSource(oToken), Left: exp, Right: exp2}
			case lexer.Less:
				exp = &LessExpression{nodeSource: makeNodeSource(oToken), Left: exp, Right: exp2}
			case lexer.LessOrEqual:
				exp = &LessOrEqualExpression{nodeSource: makeNodeSource(oToken), Left: exp, Right: exp2}
			case lexer.Greater:
				exp = &GreaterExpression{nodeSource: makeNodeSource(oToken), Left: exp, Right: exp2}
			case lexer.GreaterOrEqual:
				exp = &GreaterOrEqualExpression{nodeSource: makeNodeSource(oToken), Left: exp, Right: exp2}
			case lexer.And:
				exp = &AndExpression{nodeSource: makeNodeSource(oToken), Left: exp, Right: exp2}
			case lexer.Or:
				exp = &OrExpression{nodeSource: makeNodeSource(oToken), Left: exp, Right: exp2}
			default:
				return nil, unexpectedTokenError(oToken, lexer.Multiply, lexer.Divide, lexer.Add, lexer.Subtract,
					lexer.Equal, lexer.NotEqual, lexer.Less, lexer.LessOrEqual, lexer.Greater, lexer.GreaterOrEqual,
					lexer.And, lexer.Or)
			}
		} else {
			if pToken.Type() == lexer.LeftParenthesis && isCallableExp(exp) {
				p.getNextToken()
				var err error
				exp, err = p.parseFunctionCallExpression(pToken, exp, currentScope)
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

func (p *Parser) parseParenthesizedExpression(currentScope Scope) (Expression, error) {
	exp, err := p.parseExpression(0, currentScope)
	if err != nil {
		return nil, err
	}

	token := p.getNextToken()
	if token.Type() != lexer.RightParenthesis {
		return nil, unexpectedTokenError(token, lexer.RightParenthesis)
	}

	return exp, nil
}

func (p *Parser) parseFunctionCallExpression(startToken lexer.Token, callSource Expression, currentScope Scope) (*FunctionCallExpression, error) {
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

		exp, err := p.parseExpression(0, currentScope)
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
