package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/pbnjay/lollipops/data"
	"github.com/pbnjay/lollipops/drawing"
)

var (
	uniprot = flag.String("U", "", "Uniprot accession instead of GENE_SYMBOL")
	output  = flag.String("o", "", "output SVG file (default GENE_SYMBOL.svg)")
	width   = flag.Int("w", 0, "SVG output width (default automatic fit labels)")

	arialPath *string
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] GENE_SYMBOL [PROTEIN CHANGES ...]\n\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "Where options are:")
		flag.PrintDefaults()
	}

	if !drawing.FontLoaded() {
		arialPath = flag.String("f", "", "path to arial.ttf")
	}

	flag.Parse()

	if *arialPath != "" {
		err := drawing.LoadFontPath(*arialPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}

	if !drawing.FontLoaded() {
		fmt.Fprintln(os.Stderr, "can't find arial.ttf - for more accurate font sizing use -f=/path/to/arial.ttf")
	}

	var err error
	varStart := 0
	acc := ""
	geneSymbol := ""
	if *uniprot == "" && flag.NArg() > 0 {
		geneSymbol = flag.Arg(0)
		varStart = 1

		fmt.Fprintln(os.Stderr, "HGNC Symbol: ", flag.Arg(0))

		acc, err = data.GetProtID(flag.Arg(0))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		fmt.Fprintln(os.Stderr, "Uniprot/SwissProt Accession: ", acc)
	}

	if *uniprot != "" {
		acc = *uniprot
	}

	if flag.NArg() == 0 && *uniprot == "" {
		flag.Usage()
		os.Exit(1)
	}

	data, err := data.GetPfamGraphicData(acc)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if geneSymbol == "" {
		geneSymbol = data.Metadata.Identifier
		fmt.Fprintln(os.Stderr, "Pfam Symbol: ", geneSymbol)
	}

	if *output == "" {
		*output = geneSymbol + ".svg"
	}

	f, err := os.OpenFile(*output, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer f.Close()

	fmt.Fprintln(os.Stderr, "Drawing diagram to", *output)
	drawing.DrawSVG(f, *width, flag.Args()[varStart:], data)
}
