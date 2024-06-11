package main

import (
	"os"

	"github.com/lqqyt2423/go-monkey/repl"
)

func main() {
	repl.Start(os.Stdin, os.Stdout)
}
