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
}

type StructType struct {
	Fields []Field
}

type FunctionType struct {
	Parameters  []Field
	ReturnTypes []Type
}
