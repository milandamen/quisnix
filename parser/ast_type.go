package parser

type BasicDataType int

const (
	NoneDataType BasicDataType = iota
	IntDataType
	ByteDataType
	StringDataType
	BoolDataType
)

type BasicType struct {
	DataType BasicDataType
	Name     string
}

type StructType struct {
	Fields []Field
	Name   string
}

type FunctionType struct {
	Parameters  []*Field
	ReturnTypes []*Field // Only the type declaration is used.
}

// Type used when the definition of the time is currently unknown
type UnknownType struct {
	Name       string     // Identifier of the type that was used.
	Scope      Scope      // Scope of the place where the type was used.
	nodeSource nodeSource // Place where this identifier was used.
}

func (t BasicType) TypeName() string {
	return t.Name
}

func (t StructType) TypeName() string {
	return t.Name
}

func (t FunctionType) TypeName() string {
	return "func" // FIXME: output parameters and return types?
}

func (t UnknownType) TypeName() string {
	return t.Name
}
