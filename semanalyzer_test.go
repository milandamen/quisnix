package quisnix

import (
	"bytes"

	"github.com/milandamen/quisnix/semanalyzer"

	"github.com/milandamen/quisnix/lexer"
	"github.com/milandamen/quisnix/parser"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Semantic analyzer", func() {
	It("should fail an empty file", func() {
		a := semanalyzer.SemAnalyzer{}
		_, err := a.Analyze([]parser.Declaration{}, &parser.BuiltInScope{})
		Expect(err).ToNot(Succeed())
		Expect(err.Error()).To(Equal("must have a 'main' function"))
	})
	It("should parse a simple program", func() {
		l := lexer.Lexer{}
		p := parser.Parser{}
		a := semanalyzer.SemAnalyzer{}

		program := `
func main() {
	var a Int;
	a = 123;
	a = test(a);
	a = test(123);
}

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
		Expect(len(tokens)).To(Equal(83))

		declarations, fileScope, err := p.Parse(tokens)
		Expect(err).To(Succeed())
		Expect(len(declarations)).To(Equal(2))

		mainFunc, err := a.Analyze(declarations, fileScope)
		Expect(err).To(Succeed())
		Expect(mainFunc).ToNot(BeNil())
	})
})
