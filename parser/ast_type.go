package parser

type BasicDataType int

const (
	NoneDataType = iota
	IntDataType
	ByteDataType
	StringDataType
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
	Parameters  []Field
	ReturnTypes []Type
}

// Type used when the definition of the time is currently unknown
type UnknownType struct {
	Name string
}

func (t BasicType) TypeName() string {
	return t.Name
}

func (t StructType) TypeName() string {
	return t.Name
}

func (t FunctionType) TypeName() string {
	return "func"
}

func (t UnknownType) TypeName() string {
	return t.Name
}
