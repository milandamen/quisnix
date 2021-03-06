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

	DeclareVariable(identifier string, declaration *VariableDeclaration)
	DeclareType(identifier string, declaration *TypeDeclaration)
	DeclareFunction(identifier string, declaration *FunctionDeclaration)

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

var cachedBuiltInScopeTypes map[string]*TypeDeclaration

type BuiltInScope struct{}

func NewBuiltInScope() *BuiltInScope {
	return &BuiltInScope{}
}

type FileScope struct {
	BasicScope
	// FIXME: Imported packages scope

	// List of all top-level and sub-level type declarations in this file.
	// FIXME: move to packages scope
	AllTypeDeclarations []*TypeDeclaration

	// Note every sub-scope declaration in this file scope so that declaration
	// clashes can be found when a file-scope declaration is done after a sub-scope
	// declaration might have been done already.
	//
	// Mapped by identifier and the value is the place where the same identifier was
	// declared in a sub-scope.
	// FIXME: move to packages scope
	subScopeDeclarations map[string]nodeSource
}

func NewFileScope(parentScope Scope) *FileScope {
	return &FileScope{
		BasicScope:           *NewBasicScope(parentScope, FileScopeType),
		subScopeDeclarations: make(map[string]nodeSource),
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
		variableDeclarations: make(map[string]*VariableDeclaration),
		typeDeclarations:     make(map[string]*TypeDeclaration),
		functionDeclarations: make(map[string]*FunctionDeclaration),

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
			if currentScope.ScopeType() != FileScopeType && currentScope.ScopeType() != BuiltInScopeType {
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
			if currentScope.ScopeType() != FileScopeType && currentScope.ScopeType() != BuiltInScopeType {
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
			if currentScope.ScopeType() != FileScopeType && currentScope.ScopeType() != BuiltInScopeType {
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

func (s *BasicScope) SearchDeclaration(identifier string) Declaration {
	if decl := s.SearchVariableDeclaration(identifier); decl != nil {
		return decl
	}
	if decl := s.SearchTypeDeclaration(identifier); decl != nil {
		return decl
	}
	if decl := s.SearchFunctionDeclaration(identifier); decl != nil {
		return decl
	}

	return nil
}

func (s *BasicScope) DeclareVariable(identifier string, declaration *VariableDeclaration) {
	s.variableDeclarations[identifier] = declaration

	var currentScope Scope
	currentScope = s
	for currentScope != nil {
		if currentScope.ScopeType() == FileScopeType {
			fs, ok := currentScope.(*FileScope)
			if !ok {
				return // impossible situation as a scope of FileScopeType should always be an instance of FileScope.
			}

			if _, ok := fs.subScopeDeclarations[identifier]; !ok {
				fs.subScopeDeclarations[identifier] = declaration.nodeSource
			}
			return
		}

		currentScope = currentScope.GetParentScope()
	}
}

func (s *BasicScope) DeclareType(identifier string, declaration *TypeDeclaration) {
	s.typeDeclarations[identifier] = declaration

	var currentScope Scope
	currentScope = s
	for currentScope != nil {
		if currentScope.ScopeType() == FileScopeType {
			fs, ok := currentScope.(*FileScope)
			if !ok {
				return // impossible situation as a scope of FileScopeType should always be an instance of FileScope.
			}

			fs.AllTypeDeclarations = append(fs.AllTypeDeclarations, declaration)
			if _, ok := fs.subScopeDeclarations[identifier]; !ok {
				fs.subScopeDeclarations[identifier] = declaration.nodeSource
			}
			return
		}

		currentScope = currentScope.GetParentScope()
	}
}

func (s *BasicScope) DeclareFunction(identifier string, declaration *FunctionDeclaration) {
	s.functionDeclarations[identifier] = declaration

	var currentScope Scope
	currentScope = s
	for currentScope != nil {
		if currentScope.ScopeType() == FileScopeType {
			fs, ok := currentScope.(*FileScope)
			if !ok {
				return // impossible situation as a scope of FileScopeType should always be an instance of FileScope.
			}

			if _, ok := fs.subScopeDeclarations[identifier]; !ok {
				fs.subScopeDeclarations[identifier] = declaration.nodeSource
			}
			return
		}

		currentScope = currentScope.GetParentScope()
	}
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
		BasicScope: BasicScope{
			variableDeclarations: varDecls,
			typeDeclarations:     typeDecls,
			functionDeclarations: funcDecls,
			parentScope:          s.parentScope,
			scopeType:            s.scopeType,
		},
		subScopeDeclarations: make(map[string]nodeSource),
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

func (b *BuiltInScope) SearchDeclaration(identifier string) Declaration {
	if decl := b.SearchVariableDeclaration(identifier); decl != nil {
		return decl
	}
	if decl := b.SearchTypeDeclaration(identifier); decl != nil {
		return decl
	}
	if decl := b.SearchFunctionDeclaration(identifier); decl != nil {
		return decl
	}

	return nil
}

func (b *BuiltInScope) DeclareVariable(string, *VariableDeclaration) {
	panic("cannot declare variable on built-in scope")
}

func (b *BuiltInScope) DeclareType(string, *TypeDeclaration) {
	panic("cannot declare type on built-in scope")
}

func (b *BuiltInScope) DeclareFunction(string, *FunctionDeclaration) {
	panic("cannot declare function on built-in scope")
}

func (b *BuiltInScope) GetVariableDeclaration(string) *VariableDeclaration {
	return nil
}

func (b *BuiltInScope) GetTypeDeclaration(identifier string) *TypeDeclaration {
	if len(cachedBuiltInScopeTypes) == 0 {
		cachedBuiltInScopeTypes = map[string]*TypeDeclaration{
			"Int": {
				Type: BasicType{
					DataType: IntDataType,
					Name:     "Int",
				},
			},
			"Byte": {
				Type: BasicType{
					DataType: ByteDataType,
					Name:     "Byte",
				},
			},
			"String": {
				Type: BasicType{
					DataType: StringDataType,
					Name:     "String",
				},
			},
			"Bool": {
				Type: BasicType{
					DataType: BoolDataType,
					Name:     "Bool",
				},
			},
		}
	}

	decl, ok := cachedBuiltInScopeTypes[identifier]
	if !ok {
		return nil
	}

	return decl
}

func (b *BuiltInScope) GetFunctionDeclaration(string) *FunctionDeclaration {
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
