package quisnix

import (
	"bytes"

	"github.com/milandamen/quisnix/lexer"
	"github.com/milandamen/quisnix/parser"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Parser", func() {
	It("should parse an empty token list", func() {
		p := parser.Parser{}
		_, err := p.Parse([]lexer.Token{})
		Expect(err).To(Succeed())
	})
	It("should parse a simple program", func() {
		l := lexer.Lexer{}
		p := parser.Parser{}

		program := `
func test(asd Int) Int {
	var a Int;
	var b Byte;
	var cc String;
	a = 123 + 4;
	b = 'b';
	cc = "abc";
	a -= 2 + 3 * 4;
	a++;
	a = a + asd;
	return a;
}
`
		tokens, err := l.Parse(bytes.NewBufferString(program))
		Expect(err).To(Succeed())
		Expect(len(tokens)).To(Equal(55))

		declarations, err := p.Parse(tokens)
		Expect(err).To(Succeed())
		Expect(len(declarations)).To(Equal(1))

		testFunc := expectFunctionDeclaration(declarations[0])
		Expect(testFunc.UFSourceLine()).To(Equal(2))
		Expect(testFunc.UFSourceColumn()).To(Equal(1))
		testFuncDef := testFunc.FunctionDefinition
		testFuncType := testFuncDef.FunctionType
		Expect(len(testFuncType.Parameters)).To(Equal(1))
		Expect(len(testFuncType.ReturnTypes)).To(Equal(1))

		Expect(testFuncType.Parameters[0].Name).To(Equal("asd"))
		varASDDecl := testFuncType.Parameters[0].VariableDeclaration
		Expect(varASDDecl.UFSourceLine()).To(Equal(2))
		Expect(varASDDecl.UFSourceColumn()).To(Equal(11))
		typeDeclaration := varASDDecl.TypeDeclaration
		expectTypeDeclaration(typeDeclaration, "Int", parser.IntDataType)

		Expect(testFuncType.ReturnTypes[0].Name).To(Equal(""))
		typeDeclaration = testFuncType.ReturnTypes[0].VariableDeclaration.TypeDeclaration
		expectTypeDeclaration(typeDeclaration, "Int", parser.IntDataType)

		Expect(len(testFuncDef.Statements)).To(Equal(10))
		stmt := testFuncDef.Statements[0]
		Expect(stmt.UFSourceLine()).To(Equal(3))
		Expect(stmt.UFSourceColumn()).To(Equal(2))
		varADecl := stmt.(*parser.VariableDeclaration)
		Expect(varADecl.UFSourceLine()).To(Equal(3))
		Expect(varADecl.UFSourceColumn()).To(Equal(2))
		Expect(varADecl.DeclarationType()).To(Equal("variable"))
		expectTypeDeclaration(varADecl.TypeDeclaration, "Int", parser.IntDataType)

		stmt = testFuncDef.Statements[1]
		varBDecl := stmt.(*parser.VariableDeclaration)
		Expect(varBDecl.DeclarationType()).To(Equal("variable"))
		expectTypeDeclaration(varBDecl.TypeDeclaration, "Byte", parser.ByteDataType)

		stmt = testFuncDef.Statements[2]
		varCCDecl := stmt.(*parser.VariableDeclaration)
		Expect(varCCDecl.DeclarationType()).To(Equal("variable"))
		expectTypeDeclaration(varCCDecl.TypeDeclaration, "String", parser.StringDataType)

		stmt = testFuncDef.Statements[3]
		assignStmt := stmt.(*parser.AssignStatement)
		Expect(assignStmt.UFSourceLine()).To(Equal(6))
		Expect(assignStmt.UFSourceColumn()).To(Equal(2))
		Expect(assignStmt.VariableDeclaration).To(Equal(varADecl))
		addExp := assignStmt.Expression.(*parser.AddExpression)
		expectIntLiteralExpression(addExp.Left, 123)
		expectIntLiteralExpression(addExp.Right, 4)

		stmt = testFuncDef.Statements[4]
		assignStmt = stmt.(*parser.AssignStatement)
		Expect(assignStmt.VariableDeclaration).To(Equal(varBDecl))
		expectCharLiteralExpression(assignStmt.Expression, 'b')

		stmt = testFuncDef.Statements[5]
		assignStmt = stmt.(*parser.AssignStatement)
		Expect(assignStmt.VariableDeclaration).To(Equal(varCCDecl))
		expectStringLiteralExpression(assignStmt.Expression, "abc")

		stmt = testFuncDef.Statements[6]
		subAssignStmt := stmt.(*parser.SubtractAssignStatement)
		Expect(subAssignStmt.VariableDeclaration).To(Equal(varADecl))
		addExp = subAssignStmt.Expression.(*parser.AddExpression)
		mulExp := addExp.Right.(*parser.MultiplyExpression)
		expectIntLiteralExpression(addExp.Left, 2)
		expectIntLiteralExpression(mulExp.Left, 3)
		expectIntLiteralExpression(mulExp.Right, 4)

		stmt = testFuncDef.Statements[7]
		incStmt := stmt.(*parser.IncrementStatement)
		Expect(incStmt.VariableDeclaration).To(Equal(varADecl))

		stmt = testFuncDef.Statements[8]
		assignStmt = stmt.(*parser.AssignStatement)
		Expect(assignStmt.VariableDeclaration).To(Equal(varADecl))
		addExp = assignStmt.Expression.(*parser.AddExpression)
		expectIdentifierExpression(addExp.Left, varADecl)
		expectIdentifierExpression(addExp.Right, varASDDecl)
	})
})

func expectFunctionDeclaration(declaration parser.Declaration) *parser.FunctionDeclaration {
	d, ok := declaration.(*parser.FunctionDeclaration)
	Expect(ok).To(BeTrue())
	return d
}

func expectTypeDeclaration(declaration *parser.TypeDeclaration, typeId string, basicType parser.BasicDataType) {
	Expect(declaration.DeclarationType()).To(Equal("type"))
	typeDeclarationType := declaration.Type
	Expect(typeDeclarationType.TypeName()).To(Equal(typeId))
	Expect(typeDeclarationType.(parser.BasicType).Name).To(Equal(typeId))
	Expect(typeDeclarationType.(parser.BasicType).DataType).To(Equal(basicType))
}

func expectIntLiteralExpression(expression parser.Expression, value int) {
	exp := expression.(*parser.IntegerLiteralExpression)
	Expect(exp.Value).To(Equal(value))
}

func expectCharLiteralExpression(expression parser.Expression, value byte) {
	exp := expression.(*parser.CharacterLiteralExpression)
	Expect(exp.Value).To(Equal(value))
}

func expectStringLiteralExpression(expression parser.Expression, value string) {
	exp := expression.(*parser.StringLiteralExpression)
	Expect(exp.Value).To(Equal(value))
}

func expectIdentifierExpression(expression parser.Expression, declaration parser.Declaration) {
	exp := expression.(*parser.IdentifierExpression)
	Expect(exp.IdentifierDeclaration).To(Equal(declaration))
}
