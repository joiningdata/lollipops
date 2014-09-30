package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	uniprot = flag.String("U", "", "Uniprot accession instead of GENE_SYMBOL")
	output  = flag.String("o", "", "output SVG file (default GENE_SYMBOL.svg)")
	width   = flag.Int("w", 0, "SVG output width (default automatic fit labels)")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] GENE_SYMBOL [PROTEIN CHANGES ...]\n\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "Where options are:")
		flag.PrintDefaults()
	}
	flag.Parse()

	var err error
	varStart := 0
	acc := ""
	geneSymbol := ""
	if *uniprot == "" && flag.NArg() > 0 {
		geneSymbol = flag.Arg(0)
		varStart = 1

		fmt.Fprintln(os.Stderr, "HGNC Symbol: ", flag.Arg(0))

		acc, err = GetProtID(flag.Arg(0))
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

	data, err := GetPfamGraphicData(acc)
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
	DrawSVG(f, *width, flag.Args()[varStart:], data)
}
