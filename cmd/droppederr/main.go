package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/Orfeo42/droppederr"
)

func main() {
	singlechecker.Main(droppederr.Analyzer)
}
