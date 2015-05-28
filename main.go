package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/pbnjay/lollipops/data"
	"github.com/pbnjay/lollipops/drawing"
)

var (
	uniprot = flag.String("U", "", "Uniprot accession instead of GENE_SYMBOL")
	output  = flag.String("o", "", "output SVG/PNG file (default GENE_SYMBOL.svg)")
	width   = flag.Int("w", 0, "output width (default automatic fit labels)")
	dpi     = flag.Float64("dpi", 72, "output DPI for PNG rasterization")

	showLabels     = flag.Bool("labels", false, "draw mutation labels above lollipops")
	hideDisordered = flag.Bool("hide-disordered", false, "do not draw disordered regions")
	hideMotifs     = flag.Bool("hide-motifs", false, "do not draw motifs")
	hideAxis       = flag.Bool("hide-axis", false, "do not draw the aa position axis")
	noPatterns     = flag.Bool("no-patterns", false, "use solid fill instead of patterns for SVG output")

	synColor = flag.String("syn-color", "#0000ff", "color to use for synonymous lollipops")
	mutColor = flag.String("mut-color", "#ff0000", "color to use for non-synonymous lollipops")

	arialPath *string
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] {-U UNIPROT_ID | GENE_SYMBOL} [PROTEIN CHANGES ...]\n", os.Args[0])
		fmt.Fprintln(os.Stderr, `
Where GENE_SYMBOL is the official human HGNC gene symbol. This will use the
BioMart API to lookup the UNIPROT_ID. To skip the lookup or use other species,
specify the UniProt ID with -U (e.g. "-U P04637" for TP53)

Protein changes:
  Currently only point mutations are supported, and may be specified as:

    <AMINO><CODON><AMINO><#COLOR><@COUNT>

  Only CODON is required, and AMINO tags are not parsed.

  Synonymous mutations are denoted if the first AMINO tag matches the second
  AMINO tag, or if the second tag is not present. Otherwise the non-synonymous
  mutation color is used. The COLOR tag will override using the #RRGGBB style
  provided. The COUNT tag can be used to scale the lollipop marker size so that
  the area is exponentially proportional to the count indicated. Examples:

    R273C            -- non-synonymous mutation at codon 273
    T125@5           -- synonymous mutation at codon 125 with "5x" marker sizing
    R248Q#00ff00     -- green lollipop at codon 248
    R248Q#00ff00@131 -- green lollipop at codon 248 with "131x" marker sizing

  (N.B. color must come before count in tags)

Diagram generation options:
  -syn-color="#0000ff"    color to use for synonymous mutation markers
  -mut-color="#ff0000"    color to use for non-synonymous mutation markers
  -hide-axis              do not draw the amino position x-axis
  -hide-disordered        do not draw disordered regions on the backbone
  -hide-motifs            do not draw simple motif regions
  -labels                 draw label text above lollipop markers
  -no-patterns            use solid fill instead of patterns (SVG only)

Output options:
  -o=filename.png         set output filename (.png or .svg supported)
  -w=700                  set diagram pixel width (default = automatic fit)
  -dpi=300                set DPI (PNG output only)
`)
	}

	if !drawing.FontLoaded() {
		arialPath = flag.String("f", "", "path to arial.ttf")
	}

	flag.Parse()
	drawing.DefaultSettings.ShowLabels = *showLabels
	drawing.DefaultSettings.HideDisordered = *hideDisordered
	drawing.DefaultSettings.HideMotifs = *hideMotifs
	drawing.DefaultSettings.HideAxis = *hideAxis
	drawing.DefaultSettings.SolidFillOnly = *noPatterns
	drawing.DefaultSettings.SynonymousColor = *synColor
	drawing.DefaultSettings.MutationColor = *mutColor
	drawing.DefaultSettings.GraphicWidth = float64(*width)

	if arialPath != nil && *arialPath != "" {
		err := drawing.LoadFontPath(*arialPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}

	if !drawing.FontLoaded() {
		fmt.Fprintln(os.Stderr, "WARNING: unable to find Arial.ttf - for more accurate font sizing add -f=/path/to/arial.ttf")
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
	if strings.HasSuffix(strings.ToLower(*output), ".png") {
		drawing.DrawPNG(f, *dpi, flag.Args()[varStart:], data)
	} else {
		drawing.DrawSVG(f, flag.Args()[varStart:], data)
	}

}
