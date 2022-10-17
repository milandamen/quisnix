package quisnix

import (
	"bytes"
	"fmt"

	"github.com/milandamen/quisnix/lexer"
	"github.com/milandamen/quisnix/semanalyzer"

	"github.com/milandamen/quisnix/parser"

	"github.com/milandamen/quisnix/printer"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Printer", func() {
	FIt("should print correct LLVM IR", func() {
		l := lexer.Lexer{}
		p := parser.Parser{}
		a := semanalyzer.SemAnalyzer{}
		pr := printer.LLVMPrinter{}

		program := `
func main() Int {
	var a Int;
	a = 123;
	a = test(a);
	return a;
}

func test(asd Int) Int {
	return 2 + asd;
}
`
		tokens, err := l.Parse(bytes.NewBufferString(program))
		Expect(err).To(Succeed())
		//Expect(len(tokens)).To(Equal(20))

		declarations, fileScope, err := p.Parse(tokens)
		Expect(err).To(Succeed())
		//Expect(len(declarations)).To(Equal(1))

		mainFunc, err := a.Analyze(declarations, fileScope)
		Expect(err).To(Succeed())
		Expect(mainFunc).ToNot(BeNil())

		b := bytes.Buffer{}
		Expect(pr.Print(&b, declarations)).To(Succeed())
		fmt.Println(b.String())
	})
	PIt("should print correct LLVM IR", func() {
		l := lexer.Lexer{}
		p := parser.Parser{}
		a := semanalyzer.SemAnalyzer{}
		pr := printer.LLVMPrinter{}

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

		b := bytes.Buffer{}
		Expect(pr.Print(&b, declarations)).To(Succeed())
		fmt.Println(b.String())
	})
})
