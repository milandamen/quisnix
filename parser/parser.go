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
	unknownVarFuncIdentifiers   []*IdentifierExpression
	unknownIdentifierStatements []Statement
}

func (p *Parser) Parse(tokens []lexer.Token) ([]Declaration, *FileScope, error) {
	p.tokens = tokens
	p.tokenPos = 0

	topLevelDeclarations := make([]Declaration, 0)
	builtInScope := NewBuiltInScope()
	fileScope := NewFileScope(builtInScope)

	for true {
		tln, err := p.parseTopLevel(fileScope)
		if err != nil {
			return nil, nil, err
		}
		if tln == nil {
			break
		}

		topLevelDeclarations = append(topLevelDeclarations, tln)
	}

	if err := p.resolveUnknownTypes(); err != nil {
		return nil, nil, errors.Wrap(err, "could not resolve unknown types")
	}

	return topLevelDeclarations, fileScope, nil
}

func (p *Parser) parseTopLevel(currentScope *FileScope) (Declaration, error) {
	token := p.getNextToken()
	if token == nil {
		return nil, nil
	}

	tokenType := token.Type()
	switch tokenType {
	case lexer.Func:
		decl, err := p.parseTopLevelFunctionDeclaration(token, currentScope)
		return decl, errors.Wrapf(err, "could not parse function declaration at line %d column %d", token.UFLine(), token.UFColumn())
	default:
		return nil, unexpectedTokenError(token, lexer.Func)
	}
}

func (p *Parser) parseTopLevelFunctionDeclaration(startToken lexer.Token, currentScope *FileScope) (Declaration, error) {
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

	id := idToken.Identifier()
	ns := makeNodeSource(startToken)
	if d := currentScope.SearchDeclaration(id); d != nil {
		return nil, alreadyDeclaredError(d, ns)
	}
	if ssns, ok := currentScope.subScopeDeclarations[id]; ok {
		return nil, alreadyDeclaredInFile(ns, ssns)
	}

	def, err := p.parseFunctionDefinition(currentScope)
	if err != nil {
		return nil, err
	}

	decl := &FunctionDeclaration{
		nodeSource:         ns,
		FunctionDefinition: def,
		Name:               id,
	}

	currentScope.DeclareFunction(id, decl)
	return decl, nil
}

func (p *Parser) parseFunctionDefinition(currentScope Scope) (*FunctionDefinition, error) {
	var funcScope Scope = NewBasicScope(currentScope, FunctionScopeType)
	var parameters []*Field
	var err error
	parameters, funcScope, err = p.parseFunctionParameters(funcScope)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse function parameters")
	}

	token := p.peekNextToken()
	if token == nil {
		return nil, unexpectedEOF()
	}

	var returnTypes []*Field
	returnTypes, funcScope, err = p.parseFunctionReturnTypes(funcScope)
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

	statements, err := p.parseStatements(funcScope)
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

func (p *Parser) parseFunctionParameters(currentScope Scope) ([]*Field, Scope, error) {
	token := p.getNextToken()
	if token == nil {
		return nil, currentScope, unexpectedEOF()
	}
	if token.Type() != lexer.LeftParenthesis {
		return nil, currentScope, unexpectedTokenError(token, lexer.LeftParenthesis)
	}

	parameters := make([]*Field, 0)
	for true {
		token = p.getNextToken()
		if token == nil {
			return nil, currentScope, unexpectedEOF()
		}

		if token.Type() == lexer.RightParenthesis {
			break
		}

		if len(parameters) > 0 {
			if token.Type() != lexer.Comma {
				return nil, currentScope, unexpectedTokenError(token, lexer.RightParenthesis, lexer.Comma)
			}

			token = p.getNextToken()
			if token == nil {
				return nil, currentScope, unexpectedEOF()
			}
			if token.Type() != lexer.Identifier {
				return nil, currentScope, unexpectedTokenError(token, lexer.Identifier)
			}
		} else {
			if token.Type() != lexer.Identifier {
				return nil, currentScope, unexpectedTokenError(token, lexer.RightParenthesis, lexer.Identifier)
			}
		}

		nameToken, ok := token.(lexer.IdentifierToken)
		if !ok {
			return nil, currentScope, unexpectedTokenCastError(token)
		}

		id := nameToken.Identifier()
		if d := currentScope.SearchDeclaration(id); d != nil {
			return nil, currentScope, alreadyDeclaredError(d, nodeSource{
				line:   nameToken.Line(),
				column: nameToken.Column(),
			})
		}

		token = p.getNextToken()
		if token == nil {
			return nil, currentScope, unexpectedEOF()
		}
		if token.Type() != lexer.Identifier {
			return nil, currentScope, unexpectedTokenError(token, lexer.Identifier)
		}

		typeToken, ok := token.(lexer.IdentifierToken)
		if !ok {
			return nil, currentScope, unexpectedTokenCastError(token)
		}

		var f *Field
		var err error
		f, currentScope, err = p.getTypedField(id, typeToken, makeNodeSource(nameToken), currentScope)
		if err != nil {
			return nil, currentScope, err
		}

		parameters = append(parameters, f)
	}

	return parameters, currentScope, nil
}

func (p *Parser) parseFunctionReturnTypes(currentScope Scope) ([]*Field, Scope, error) {
	token := p.peekNextToken()
	if token == nil {
		return nil, currentScope, unexpectedEOF()
	}

	returnTypes := make([]*Field, 0)
	if token.Type() == lexer.Identifier {
		p.getNextToken()
		typeToken, ok := token.(lexer.IdentifierToken)
		if !ok {
			return nil, currentScope, unexpectedTokenCastError(token)
		}

		var f *Field
		var err error
		f, currentScope, err = p.getTypedField("", typeToken, makeNodeSource(typeToken), currentScope)
		if err != nil {
			return nil, currentScope, err
		}

		return []*Field{f}, currentScope, nil
	} else if token.Type() == lexer.LeftParenthesis {
		p.getNextToken()
	} else {
		return make([]*Field, 0), currentScope, nil
	}

	for true {
		token := p.getNextToken()
		if token == nil {
			return nil, currentScope, unexpectedEOF()
		}

		if len(returnTypes) > 0 {
			if token.Type() == lexer.RightParenthesis {
				break
			}

			if token.Type() != lexer.Comma {
				return nil, currentScope, unexpectedTokenError(token, lexer.RightParenthesis, lexer.Comma)
			}

			token = p.getNextToken()
			if token == nil {
				return nil, currentScope, unexpectedEOF()
			}
		}

		if token.Type() != lexer.Identifier {
			return nil, currentScope, unexpectedTokenError(token, lexer.Identifier)
		}

		typeToken, ok := token.(lexer.IdentifierToken)
		if !ok {
			return nil, currentScope, unexpectedTokenCastError(token)
		}

		var f *Field
		var err error
		f, currentScope, err = p.getTypedField("", typeToken, makeNodeSource(typeToken), currentScope)
		if err != nil {
			return nil, currentScope, err
		}

		returnTypes = append(returnTypes, f)
	}

	return returnTypes, currentScope, nil
}

func (p *Parser) getTypedField(fieldName string, typeToken lexer.IdentifierToken, ns nodeSource, currentScope Scope) (*Field, Scope, error) {
	// TODO what if its a complex type?
	var f *Field
	typeId := typeToken.Identifier()
	typeDecl := currentScope.SearchTypeDeclaration(typeId)

	varDecl := &VariableDeclaration{
		nodeSource:      ns,
		TypeDeclaration: typeDecl,
	}

	if typeDecl != nil {
		f = &Field{
			Name:                fieldName,
			VariableDeclaration: varDecl,
		}
	} else {
		typeDecl = &TypeDeclaration{
			nodeSource: nodeSource{},
			Type:       UnknownType{Name: typeId, Scope: currentScope, nodeSource: makeNodeSource(typeToken)},
		}

		varDecl.TypeDeclaration = typeDecl

		f = &Field{
			Name:                fieldName,
			VariableDeclaration: varDecl,
		}

		p.unknownFieldTypes = append(p.unknownFieldTypes, f)
	}

	if fieldName != "" {
		if d := currentScope.SearchDeclaration(fieldName); d != nil {
			return nil, currentScope, alreadyDeclaredError(d, ns)
		}

		currentScope = currentScope.CloneShallow()
		currentScope.DeclareVariable(fieldName, varDecl)
	}

	return f, currentScope, nil
}

func (p *Parser) parseStatements(currentScope Scope) ([]Statement, error) {
	statements := make([]Statement, 0)
	for true {
		token := p.getNextToken()
		if token == nil {
			return nil, unexpectedEOF()
		}

		switch token.Type() {
		// TODO implement For
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
			var stmt Statement
			var err error
			stmt, currentScope, err = p.parseVariableDeclarationStatement(token, currentScope)
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
		case lexer.Return:
			stmt, err := p.parseReturnStatement(token, currentScope)
			if err != nil {
				return nil, err
			}

			token = p.getNextToken()
			if token == nil {
				return nil, unexpectedEOF()
			}
			if token.Type() != lexer.RightBrace {
				return nil, unexpectedTokenError(token, lexer.RightBrace)
			}

			// Return must be the last statement in the block.
			return append(statements, stmt), nil
		case lexer.RightBrace:
			return statements, nil
		default:
			return nil, unexpectedTokenError(token, lexer.Identifier, lexer.If, lexer.For, lexer.While, lexer.Var, lexer.Return, lexer.RightBrace)
		}
	}

	return nil, errors.New("unreachable code: Parser.parseStatements after loop")
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

func (p *Parser) parseVariableDeclarationStatement(startToken lexer.Token, currentScope Scope) (Statement, Scope, error) {
	token := p.getNextToken()
	if token == nil {
		return nil, currentScope, unexpectedEOF()
	}
	if token.Type() != lexer.Identifier {
		return nil, currentScope, unexpectedTokenError(token, lexer.Identifier)
	}

	idToken, ok := token.(lexer.IdentifierToken)
	if !ok {
		return nil, currentScope, unexpectedTokenCastError(token)
	}

	id := idToken.Identifier()
	ns := makeNodeSource(startToken)
	if d := currentScope.SearchDeclaration(id); d != nil {
		return nil, currentScope, alreadyDeclaredError(d, ns)
	}

	token = p.getNextToken()
	if token == nil {
		return nil, currentScope, unexpectedEOF()
	}
	if token.Type() != lexer.Identifier {
		return nil, currentScope, unexpectedTokenError(token, lexer.Identifier)
	}

	typeToken, ok := token.(lexer.IdentifierToken)
	if !ok {
		return nil, currentScope, unexpectedTokenError(token)
	}

	typeId := typeToken.Identifier()
	var varDecl *VariableDeclaration
	decl := currentScope.SearchTypeDeclaration(typeId)
	if decl != nil {
		varDecl = &VariableDeclaration{
			nodeSource:      ns,
			TypeDeclaration: decl,
		}
	} else {
		varDecl = &VariableDeclaration{
			nodeSource: ns,
			TypeDeclaration: &TypeDeclaration{
				nodeSource: nodeSource{},
				Type:       UnknownType{Name: typeId, Scope: currentScope, nodeSource: makeNodeSource(typeToken)},
			},
		}

		p.unknownIdentifierStatements = append(p.unknownIdentifierStatements, varDecl)
	}

	currentScope = currentScope.CloneShallow()
	currentScope.DeclareVariable(id, varDecl)

	token = p.getNextToken()
	if token == nil {
		return nil, currentScope, unexpectedEOF()
	}
	if token.Type() != lexer.Semicolon {
		return nil, currentScope, unexpectedTokenError(token, lexer.Semicolon)
	}

	return varDecl, currentScope, nil
}

func (p *Parser) parseIdentifierStatement(idToken lexer.IdentifierToken, currentScope Scope) (Statement, error) {
	token := p.getNextToken()
	if token == nil {
		return nil, unexpectedEOF()
	}

	id := idToken.Identifier()
	var varDecl Declaration
	var addUnknownIdentifierStmt bool
	if d := currentScope.SearchVariableDeclaration(id); d != nil {
		varDecl = d
	} else {
		addUnknownIdentifierStmt = true
		varDecl = &UnknownDeclaration{
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
			VariableDeclaration: varDecl,
			Expression:          exp,
		}
	case lexer.AddAssign:
		exp, err := p.parseExpression(0, currentScope)
		if err != nil {
			return nil, err
		}

		stmt = &AddAssignStatement{
			nodeSource:          makeNodeSource(idToken),
			VariableDeclaration: varDecl,
			Expression:          exp,
		}
	case lexer.SubtractAssign:
		exp, err := p.parseExpression(0, currentScope)
		if err != nil {
			return nil, err
		}

		stmt = &SubtractAssignStatement{
			nodeSource:          makeNodeSource(idToken),
			VariableDeclaration: varDecl,
			Expression:          exp,
		}
	case lexer.Increment:
		// TODO check that varDecl has type Int
		stmt = &IncrementStatement{
			nodeSource:          makeNodeSource(idToken),
			VariableDeclaration: varDecl,
		}
	case lexer.Decrement:
		stmt = &DecrementStatement{
			nodeSource:          makeNodeSource(idToken),
			VariableDeclaration: varDecl,
		}
	case lexer.LeftParenthesis:
		var addUnknownIdentifierExp bool
		if addUnknownIdentifierStmt {
			addUnknownIdentifierStmt = false
			funcDecl := currentScope.SearchFunctionDeclaration(id)
			if funcDecl != nil {
				varDecl = funcDecl
			} else {
				addUnknownIdentifierExp = true
			}
		}

		exp := newIdentifierExpression(makeNodeSource(idToken), varDecl)

		if addUnknownIdentifierExp {
			p.unknownVarFuncIdentifiers = append(p.unknownVarFuncIdentifiers, exp)
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

func (p *Parser) parseReturnStatement(startToken lexer.Token, currentScope Scope) (Statement, error) {
	exps := make([]Expression, 0)
	for true {
		token := p.peekNextToken()
		if token == nil {
			return nil, unexpectedEOF()
		}
		if token.Type() == lexer.Semicolon {
			p.getNextToken()
			break
		}

		if len(exps) > 0 {
			if token.Type() != lexer.Comma {
				return nil, unexpectedTokenError(token, lexer.Semicolon, lexer.Comma)
			}

			p.getNextToken()
		}

		exp, err := p.parseExpression(0, currentScope)
		if err != nil {
			return nil, err
		}

		exps = append(exps, exp)
	}

	return &ReturnStatement{
		nodeSource:        makeNodeSource(startToken),
		ReturnExpressions: exps,
	}, nil
}

func (p *Parser) parseExpression(prevOperatorPrecedence int, currentScope Scope) (Expression, error) {
	token := p.getNextToken()
	if token == nil {
		return nil, unexpectedEOF()
	}

	var notExp *NotExpression
	ns := makeNodeSource(token)
	if token.Type() == lexer.Not {
		notExp = newNotExpression(ns, nil, currentScope)
	}

	var exp Expression
	switch token.Type() {
	case lexer.Integer:
		intToken, ok := token.(lexer.IntegerToken)
		if !ok {
			return nil, unexpectedTokenCastError(token)
		}

		exp = newIntegerLiteralExpression(ns, intToken.Integer(), currentScope)
	case lexer.Character:
		charToken, ok := token.(lexer.CharacterToken)
		if !ok {
			return nil, unexpectedTokenCastError(token)
		}

		exp = newCharacterLiteralExpression(ns, charToken.Character(), currentScope)
	case lexer.String:
		stringToken, ok := token.(lexer.StringToken)
		if !ok {
			return nil, unexpectedTokenCastError(token)
		}

		exp = newStringLiteralExpression(ns, stringToken.String(), currentScope)
	case lexer.True:
		exp = newBooleanLiteralExpression(ns, true, currentScope)
	case lexer.False:
		exp = newBooleanLiteralExpression(ns, false, currentScope)
	case lexer.Identifier:
		idToken, ok := token.(lexer.IdentifierToken)
		if !ok {
			return nil, unexpectedTokenCastError(token)
		}

		id := idToken.Identifier()
		var decl Declaration
		if d := currentScope.SearchVariableDeclaration(id); d == nil {
			if d2 := currentScope.SearchFunctionDeclaration(id); d2 == nil {
				decl = &UnknownDeclaration{
					nodeSource: makeNodeSource(idToken),
					Identifier: id,
					Scope:      currentScope,
				}
				idExp := newIdentifierExpression(ns, decl)

				p.unknownVarFuncIdentifiers = append(p.unknownVarFuncIdentifiers, idExp)
				exp = idExp
				break
			} else {
				decl = d2
			}
		} else {
			decl = d
		}

		exp = newIdentifierExpression(ns, decl)
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
		// TODO implement function-as-first-citizen calling.
		//if _, ok := e.(*FunctionCallExpression); ok {
		//	return true
		//}
		return false
	}

	for true {
		pToken := p.peekNextToken()
		if pToken == nil {
			break
		}

		oToken, ok := pToken.(lexer.OperatorToken)
		if ok {
			precedence := oToken.OperatorPrecedence()
			if precedence <= prevOperatorPrecedence {
				break
			}

			p.getNextToken()
			exp2, err := p.parseExpression(precedence, currentScope)
			if err != nil {
				return nil, err
			}

			source := makeNodeSource(oToken)
			switch oToken.Type() {
			case lexer.Multiply:
				exp = newMultiplyExpression(source, exp, exp2)
			case lexer.Divide:
				exp = newDivideExpression(source, exp, exp2)
			case lexer.Add:
				exp = newAddExpression(source, exp, exp2)
			case lexer.Subtract:
				exp = newSubtractExpression(source, exp, exp2)
			case lexer.Equal:
				exp = newEqualExpression(source, exp, exp2, currentScope)
			case lexer.NotEqual:
				exp = newNotEqualExpression(source, exp, exp2, currentScope)
			case lexer.Less:
				exp = newLessExpression(source, exp, exp2, currentScope)
			case lexer.LessOrEqual:
				exp = newLessOrEqualExpression(source, exp, exp2, currentScope)
			case lexer.Greater:
				exp = newGreaterExpression(source, exp, exp2, currentScope)
			case lexer.GreaterOrEqual:
				exp = newGreaterOrEqualExpression(source, exp, exp2, currentScope)
			case lexer.And:
				exp = newAndExpression(source, exp, exp2, currentScope)
			case lexer.Or:
				exp = newOrExpression(source, exp, exp2, currentScope)
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
		pToken := p.peekNextToken()
		if pToken == nil {
			return nil, unexpectedEOF()
		}

		if pToken.Type() == lexer.RightParenthesis {
			p.getNextToken()
			break
		}

		if len(parameters) > 0 {
			p.getNextToken()
			if pToken.Type() != lexer.Comma {
				return nil, unexpectedTokenError(pToken, lexer.RightParenthesis, lexer.Comma)
			}
		}

		exp, err := p.parseExpression(0, currentScope)
		if err != nil {
			return nil, err
		}

		parameters = append(parameters, exp)
	}

	return newFunctionCallExpression(makeNodeSource(startToken), callSource, parameters), nil
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
	if p.tokenPos >= len(p.tokens) {
		return nil
	}

	return p.tokens[p.tokenPos]
}

func (p *Parser) resolveUnknownTypes() error {
	for _, f := range p.unknownFieldTypes {
		t := f.VariableDeclaration.TypeDeclaration.Type
		if ut, ok := t.(UnknownType); ok {
			typeId := ut.Name
			decl := ut.Scope.SearchTypeDeclaration(typeId)
			if decl != nil {
				f.VariableDeclaration.TypeDeclaration = decl
			} else {
				return errors.Errorf("no type found for identifier '%s' at line %d column %d",
					typeId, ut.nodeSource.UFSourceLine(), ut.nodeSource.UFSourceColumn())
			}
		} else {
			return errors.New("error resolving unknownFieldTypes: expected type of TypeDeclaration.Type to be UnknownType")
		}
	}

	for _, exp := range p.unknownVarFuncIdentifiers {
		idDecl := exp.IdentifierDeclaration
		if d, ok := idDecl.(*UnknownDeclaration); ok {
			id := d.Identifier

			var decl Declaration
			if dv := d.Scope.SearchVariableDeclaration(id); dv == nil {
				if df := d.Scope.SearchFunctionDeclaration(id); df == nil {
					return errors.Errorf("no variable or function found for identifier '%s' at line %d column %d",
						id, exp.UFSourceLine(), exp.UFSourceColumn())
				} else {
					decl = df
				}
			} else {
				decl = dv
			}

			exp.IdentifierDeclaration = decl
		} else {
			return errors.New("error resolving unknownVarFuncIdentifiers: expected type of IdentifierExpression.IdentifierDeclaration to be UnknownDeclaration")
		}
	}

	for _, stmt := range p.unknownIdentifierStatements {
		if varDecl, ok := stmt.(*VariableDeclaration); ok {
			t := varDecl.TypeDeclaration.Type
			if ut, ok := t.(UnknownType); ok {
				typeId := ut.Name
				decl := ut.Scope.SearchTypeDeclaration(typeId)
				if decl != nil {
					varDecl.TypeDeclaration = decl
				} else {
					return errors.Errorf("no type found for identifier '%s' at line %d column %d",
						typeId, ut.nodeSource.UFSourceLine(), ut.nodeSource.UFSourceColumn())
				}
			} else {
				return errors.New("error resolving unknownIdentifierStatements: expected type of VariableDeclaration.TypeDeclaration.Type to be UnknownType")
			}
		} else if s, ok := stmt.(StatementHavingVariableDeclaration); ok {
			if d, ok := s.GetVariableDeclaration().(*UnknownDeclaration); ok {
				id := d.Identifier
				decl := d.Scope.SearchVariableDeclaration(id)
				if decl != nil {
					s.SetVariableDeclaration(decl)
				} else {
					return errors.Errorf("no variable found for identifier '%s' at line %d column %d",
						id, d.UFSourceLine(), d.UFSourceColumn())
				}
			} else {
				return errors.New("error resolving unknownIdentifierStatements: expected type of stmt.VariableDeclaration to be UnknownDeclaration")
			}
		} else {
			return errors.New("error resolving unknownIdentifierStatements: unknown statement type")
		}
	}

	return nil
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

func alreadyDeclaredError(d Declaration, currentNodeSource nodeSource) error {
	t := d.DeclarationType()
	if t == "unknown" {
		return errors.Errorf("declaration at line %d column %d was already declared at line %d column %d",
			currentNodeSource.UFSourceLine(), currentNodeSource.UFSourceColumn(),
			d.UFSourceLine(), d.UFSourceColumn())
	}

	return errors.Errorf("declaration at line %d column %d was already declared as a '%s' at line %d column %d",
		currentNodeSource.UFSourceLine(), currentNodeSource.UFSourceColumn(),
		t, d.UFSourceLine(), d.UFSourceColumn())
}

func alreadyDeclaredInFile(currentNodeSource nodeSource, subScopeNodeSource nodeSource) error {
	return errors.Errorf("declaration at line %d column %d was already declared in file scope at line %d column %d",
		subScopeNodeSource.UFSourceLine(), subScopeNodeSource.UFSourceColumn(),
		currentNodeSource.UFSourceLine(), currentNodeSource.UFSourceColumn())
}
