package parser

type ScopeType int

const (
	BlockScopeType ScopeType = iota
	FunctionScopeType
	FileScopeType
	BuiltInScopeType
)

// Scope describes the declared variables, types and functions for the current code block,
// function or file.
type Scope interface {
	// Search the variable declaration belonging to the given identifier in the scope tree.
	// When no suitable variable declaration is found, nil is returned.
	SearchVariableDeclaration(identifier string) *VariableDeclaration

	// Search the type declaration belonging to the given identifier in the scope tree.
	// When no suitable type declaration is found, nil is returned.
	SearchTypeDeclaration(identifier string) *TypeDeclaration

	// Search the function declaration belonging to the given identifier in the scope tree.
	// When no suitable function declaration is found, nil is returned.
	SearchFunctionDeclaration(identifier string) *FunctionDeclaration

	// Search the declaration belonging to the given identifier in the scope tree
	SearchDeclaration(identifier string) Declaration

	DeclareVariable(identifier string, declaration VariableDeclaration)
	DeclareType(identifier string, declaration TypeDeclaration)
	DeclareFunction(identifier string, declaration FunctionDeclaration)

	// Get the variable declaration belonging to the given identifier in the current scope.
	// When no suitable variable declaration is found, nil is returned.
	GetVariableDeclaration(identifier string) *VariableDeclaration

	// Get the type declaration belonging to the given identifier in the current scope.
	// When no suitable type declaration is found, nil is returned.
	GetTypeDeclaration(identifier string) *TypeDeclaration

	// Get the function declaration belonging to the given identifier in the current scope.
	// When no suitable function declaration is found, nil is returned.
	GetFunctionDeclaration(identifier string) *FunctionDeclaration

	GetParentScope() Scope
	ScopeType() ScopeType
	CloneShallow() Scope
}

type BuiltInScope struct{}

func NewBuiltInScope() *BuiltInScope {
	return &BuiltInScope{}
}

type FileScope struct {
	BasicScope
	// FIXME: Imported packages scope

	// TODO note every sub-scope declaration in this file scope so that declaration
	//  clashes can be found when a file-scope declaration is done after a sub-scope
	//  declaration might have been done already.
}

func NewFileScope(parentScope Scope) *FileScope {
	return &FileScope{
		BasicScope: BasicScope{
			parentScope: parentScope,
			scopeType:   FileScopeType,
		},
	}
}

type BasicScope struct {
	variableDeclarations map[string]*VariableDeclaration
	typeDeclarations     map[string]*TypeDeclaration
	functionDeclarations map[string]*FunctionDeclaration

	parentScope Scope
	scopeType   ScopeType
}

func NewBasicScope(parentScope Scope, scopeType ScopeType) *BasicScope {
	return &BasicScope{
		parentScope: parentScope,
		scopeType:   scopeType,
	}
}

func (s *BasicScope) SearchVariableDeclaration(identifier string) *VariableDeclaration {
	var currentScope Scope
	currentScope = s
	skipTillTopLevel := false

	for currentScope != nil {
		if skipTillTopLevel {
			if currentScope.ScopeType() != FileScopeType {
				currentScope = currentScope.GetParentScope()
				continue
			}
		} else {
			if currentScope.ScopeType() == FunctionScopeType {
				skipTillTopLevel = true
			}
		}

		decl := currentScope.GetVariableDeclaration(identifier)
		if decl != nil {
			return decl
		}

		currentScope = currentScope.GetParentScope()
	}

	return nil
}

func (s *BasicScope) SearchTypeDeclaration(identifier string) *TypeDeclaration {
	var currentScope Scope
	currentScope = s
	skipTillTopLevel := false

	for currentScope != nil {
		if skipTillTopLevel {
			if currentScope.ScopeType() != FileScopeType {
				currentScope = currentScope.GetParentScope()
				continue
			}
		} else {
			if currentScope.ScopeType() == FunctionScopeType {
				skipTillTopLevel = true
			}
		}

		decl := currentScope.GetTypeDeclaration(identifier)
		if decl != nil {
			return decl
		}

		currentScope = currentScope.GetParentScope()
	}

	return nil
}

func (s *BasicScope) SearchFunctionDeclaration(identifier string) *FunctionDeclaration {
	var currentScope Scope
	currentScope = s
	skipTillTopLevel := false

	for currentScope != nil {
		if skipTillTopLevel {
			if currentScope.ScopeType() != FileScopeType {
				currentScope = currentScope.GetParentScope()
				continue
			}
		} else {
			if currentScope.ScopeType() == FunctionScopeType {
				skipTillTopLevel = true
			}
		}

		decl := currentScope.GetFunctionDeclaration(identifier)
		if decl != nil {
			return decl
		}

		currentScope = currentScope.GetParentScope()
	}

	return nil
}

func (s *BasicScope) DeclareVariable(identifier string, declaration VariableDeclaration) {
	s.variableDeclarations[identifier] = &declaration
}

func (s *BasicScope) DeclareType(identifier string, declaration TypeDeclaration) {
	s.typeDeclarations[identifier] = &declaration
}

func (s *BasicScope) DeclareFunction(identifier string, declaration FunctionDeclaration) {
	s.functionDeclarations[identifier] = &declaration
}

func (s *BasicScope) GetVariableDeclaration(identifier string) *VariableDeclaration {
	decl, ok := s.variableDeclarations[identifier]
	if !ok {
		return nil
	}

	return decl
}

func (s *BasicScope) GetTypeDeclaration(identifier string) *TypeDeclaration {
	decl, ok := s.typeDeclarations[identifier]
	if !ok {
		return nil
	}

	return decl
}

func (s *BasicScope) GetFunctionDeclaration(identifier string) *FunctionDeclaration {
	decl, ok := s.functionDeclarations[identifier]
	if !ok {
		return nil
	}

	return decl
}

func (s *BasicScope) GetParentScope() Scope {
	return s.parentScope
}

func (s *BasicScope) ScopeType() ScopeType {
	return s.scopeType
}

func (s *BasicScope) CloneShallow() Scope {
	varDecls := make(map[string]*VariableDeclaration)
	typeDecls := make(map[string]*TypeDeclaration)
	funcDecls := make(map[string]*FunctionDeclaration)

	for k, v := range s.variableDeclarations {
		varDecls[k] = v
	}
	for k, v := range s.typeDeclarations {
		typeDecls[k] = v
	}
	for k, v := range s.functionDeclarations {
		funcDecls[k] = v
	}

	return &BasicScope{
		variableDeclarations: varDecls,
		typeDeclarations:     typeDecls,
		functionDeclarations: funcDecls,
		parentScope:          s.parentScope,
		scopeType:            s.scopeType,
	}
}

func (s *FileScope) ScopeType() ScopeType {
	return FileScopeType
}

func (s *FileScope) CloneShallow() Scope {
	varDecls := make(map[string]*VariableDeclaration)
	typeDecls := make(map[string]*TypeDeclaration)
	funcDecls := make(map[string]*FunctionDeclaration)

	for k, v := range s.variableDeclarations {
		varDecls[k] = v
	}
	for k, v := range s.typeDeclarations {
		typeDecls[k] = v
	}
	for k, v := range s.functionDeclarations {
		funcDecls[k] = v
	}

	return &FileScope{
		BasicScope{
			variableDeclarations: varDecls,
			typeDeclarations:     typeDecls,
			functionDeclarations: funcDecls,
			parentScope:          s.parentScope,
			scopeType:            s.scopeType,
		},
	}
}

func (b *BuiltInScope) SearchVariableDeclaration(identifier string) *VariableDeclaration {
	return b.GetVariableDeclaration(identifier)
}

func (b *BuiltInScope) SearchTypeDeclaration(identifier string) *TypeDeclaration {
	return b.GetTypeDeclaration(identifier)
}

func (b *BuiltInScope) SearchFunctionDeclaration(identifier string) *FunctionDeclaration {
	return b.GetFunctionDeclaration(identifier)
}

func (b *BuiltInScope) DeclareVariable(identifier string, declaration VariableDeclaration) {
	panic("cannot declare variable on built-in scope")
}

func (b *BuiltInScope) DeclareType(identifier string, declaration TypeDeclaration) {
	panic("cannot declare type on built-in scope")
}

func (b *BuiltInScope) DeclareFunction(identifier string, declaration FunctionDeclaration) {
	panic("cannot declare function on built-in scope")
}

func (b *BuiltInScope) GetVariableDeclaration(identifier string) *VariableDeclaration {
	return nil
}

func (b *BuiltInScope) GetTypeDeclaration(identifier string) *TypeDeclaration {
	// TODO add and cache built-in types
	return nil
}

func (b *BuiltInScope) GetFunctionDeclaration(identifier string) *FunctionDeclaration {
	return nil
}

func (b *BuiltInScope) GetParentScope() Scope {
	return nil
}

func (b *BuiltInScope) ScopeType() ScopeType {
	return BuiltInScopeType
}

func (b *BuiltInScope) CloneShallow() Scope {
	return &BuiltInScope{}
}
