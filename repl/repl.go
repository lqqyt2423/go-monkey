package repl

import (
	"bufio"
	"fmt"
	"io"

	"github.com/lqqyt2423/go-monkey/lexer"
	"github.com/lqqyt2423/go-monkey/parser"
)

const PROMPT = ">> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	for {
		fmt.Fprintf(out, PROMPT)
		if !scanner.Scan() {
			return
		}
		line := scanner.Text()
		l := lexer.New(line)
		p := parser.New(l)
		program := p.ParseProgram()
		if len(p.Errors()) > 0 {
			printParseErrors(out, p.Errors())
			continue
		}
		fmt.Fprintf(out, "%s\n", program.String())
	}
}

func printParseErrors(out io.Writer, errors []string) {
	for _, msg := range errors {
		fmt.Fprintf(out, "\t%s\n", msg)
	}
}
