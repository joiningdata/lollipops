package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	output = flag.String("o", "", "output SVG file")
	width  = flag.Int("w", 740, "SVG output width")
)

func main() {
	flag.Parse()
	if *output == "" || flag.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "Usage: %s [-w 500] -o out.svg GENE_SYMBOL [CHANGE1 CHANGE2 ...]\n", os.Args[0])
		return
	}

	f, err := os.OpenFile(*output, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	fmt.Fprintln(os.Stderr, "HGNC Symbol: ", flag.Arg(0))

	acc, err := GetProtID(flag.Arg(0))
	if err != nil {
		panic(err)
	}
	fmt.Fprintln(os.Stderr, "Uniprot/SwissProt Accession: ", acc)

	data, err := GetPfamGraphicData(acc)
	if err != nil {
		panic(err)
	}

	DrawSVG(f, *width, flag.Args()[1:], data)
}
