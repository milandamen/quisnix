package quisnix_test

import (
	"bytes"

	"github.com/milandamen/quisnix"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Lexer", func() {
	It("should parse an empty string", func() {
		lexer := quisnix.Lexer{}
		Expect(lexer.Parse(bytes.NewBuffer([]byte{}))).To(Succeed())
	})

	It("should parse a simple program", func() {
		lexer := quisnix.Lexer{}
		program := `
 
	
func Test(asd int) int {
	var a = 123 + 4;
	var b = 'b';
	var cc = "abc";
	a -= 2;
	a++;
	a = a + asd;
	return a;
}`

		Expect(lexer.Parse(bytes.NewBufferString(program))).To(Succeed())
		tokens := lexer.Tokens()
		Expect(len(tokens)).To(Equal(42))

		Expect(tokens[0].Type()).To(Equal(quisnix.Func))
		Expect(tokens[0].Line()).To(Equal(3))
		Expect(tokens[0].Column()).To(Equal(0))
		expectIdentifierToken(tokens[1], "Test")
		Expect(tokens[1].Line()).To(Equal(3))
		Expect(tokens[1].Column()).To(Equal(5))
		Expect(tokens[2].Type()).To(Equal(quisnix.LeftParenthesis))
		expectIdentifierToken(tokens[3], "asd")
		expectIdentifierToken(tokens[4], "int")
		Expect(tokens[5].Type()).To(Equal(quisnix.RightParenthesis))
		expectIdentifierToken(tokens[6], "int")
		Expect(tokens[7].Type()).To(Equal(quisnix.LeftBrace))

		Expect(tokens[8].Type()).To(Equal(quisnix.Var))
		expectIdentifierToken(tokens[9], "a")
		Expect(tokens[10].Type()).To(Equal(quisnix.Assign))
		expectLiteralIntegerToken(tokens[11], 123)
		Expect(tokens[12].Type()).To(Equal(quisnix.Add))
		expectLiteralIntegerToken(tokens[13], 4)
		Expect(tokens[14].Type()).To(Equal(quisnix.Semicolon))

		Expect(tokens[15].Type()).To(Equal(quisnix.Var))
		expectIdentifierToken(tokens[16], "b")
		Expect(tokens[17].Type()).To(Equal(quisnix.Assign))
		expectLiteralCharacterToken(tokens[18], 'b')
		Expect(tokens[19].Type()).To(Equal(quisnix.Semicolon))

		Expect(tokens[20].Type()).To(Equal(quisnix.Var))
		expectIdentifierToken(tokens[21], "cc")
		Expect(tokens[22].Type()).To(Equal(quisnix.Assign))
		expectLiteralStringToken(tokens[23], "abc")
		Expect(tokens[24].Type()).To(Equal(quisnix.Semicolon))

		expectIdentifierToken(tokens[25], "a")
		Expect(tokens[26].Type()).To(Equal(quisnix.SubtractAssign))
		expectLiteralIntegerToken(tokens[27], 2)
		Expect(tokens[28].Type()).To(Equal(quisnix.Semicolon))

		expectIdentifierToken(tokens[29], "a")
		Expect(tokens[30].Type()).To(Equal(quisnix.Increment))
		Expect(tokens[31].Type()).To(Equal(quisnix.Semicolon))

		expectIdentifierToken(tokens[32], "a")
		Expect(tokens[33].Type()).To(Equal(quisnix.Assign))
		expectIdentifierToken(tokens[34], "a")
		Expect(tokens[35].Type()).To(Equal(quisnix.Add))
		expectIdentifierToken(tokens[36], "asd")
		Expect(tokens[37].Type()).To(Equal(quisnix.Semicolon))

		Expect(tokens[38].Type()).To(Equal(quisnix.Return))
		expectIdentifierToken(tokens[39], "a")
		Expect(tokens[40].Type()).To(Equal(quisnix.Semicolon))

		Expect(tokens[41].Type()).To(Equal(quisnix.RightBrace))
	})
})

func expectIdentifierToken(token quisnix.Token, identifier string) {
	Expect(token.Type()).To(Equal(quisnix.Identifier))
	t, ok := token.(quisnix.IdentifierToken)
	Expect(ok).To(BeTrue())
	Expect(t.Identifier()).To(Equal(identifier))
}

func expectLiteralIntegerToken(token quisnix.Token, integer int) {
	Expect(token.Type()).To(Equal(quisnix.Integer))
	t, ok := token.(quisnix.IntegerToken)
	Expect(ok).To(BeTrue())
	Expect(t.Integer()).To(Equal(integer))
}

func expectLiteralCharacterToken(token quisnix.Token, char byte) {
	Expect(token.Type()).To(Equal(quisnix.Character))
	t, ok := token.(quisnix.CharacterToken)
	Expect(ok).To(BeTrue())
	Expect(t.Character()).To(Equal(char))
}

func expectLiteralStringToken(token quisnix.Token, s string) {
	Expect(token.Type()).To(Equal(quisnix.String))
	t, ok := token.(quisnix.StringToken)
	Expect(ok).To(BeTrue())
	Expect(t.String()).To(Equal(s))
}
