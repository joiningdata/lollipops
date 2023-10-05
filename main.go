//
//    Lollipops command-line diagram generator for genetic variations.
//    Copyright (C) 2015 Jeremy Jay <jeremy@pbnjay.com>
//
//    This program is free software: you can redistribute it and/or modify
//    it under the terms of the GNU General Public License as published by
//    the Free Software Foundation, either version 3 of the License, or
//    (at your option) any later version.
//
//    This program is distributed in the hope that it will be useful,
//    but WITHOUT ANY WARRANTY; without even the implied warranty of
//    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//    GNU General Public License for more details.
//
//    You should have received a copy of the GNU General Public License
//    along with this program.  If not, see <http://www.gnu.org/licenses/>.

// Command lollipops provides a command-line interface for generating lollipop-style
// genetic variation diagrams. It should be suitable for various scriptable pipelines
// and other internal workflows.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/inconshreveable/mousetrap"
	"github.com/joiningdata/lollipops/data"
	"github.com/joiningdata/lollipops/drawing"
)

var (
	queryDB = flag.String("Q", "GENENAME", "Uniprot query database when -U not used")
	uniprot = flag.String("U", "", "Uniprot accession instead of GENE_SYMBOL")
	output  = flag.String("o", "", "output SVG/PNG file (default GENE_SYMBOL.svg)")
	width   = flag.Int("w", 0, "output width (default automatic fit labels)")
	dpi     = flag.Float64("dpi", 72, "output DPI for PNG rasterization")

	showLegend     = flag.Bool("legend", false, "draw a legend for colored regions")
	showLabels     = flag.Bool("labels", false, "draw mutation labels above lollipops")
	showDisordered = flag.Bool("show-disordered", false, "draw disordered regions on the backbone")
	showMotifs     = flag.Bool("show-motifs", false, "draw simple motif regions")
	hideAxis       = flag.Bool("hide-axis", false, "do not draw the aa position axis")
	noPatterns     = flag.Bool("no-patterns", false, "use solid fill instead of patterns for SVG output")
	domainLabels   = flag.String("domain-labels", "truncated", "how to apply domain labels")

	synColor = flag.String("syn-color", "#0000ff", "color to use for synonymous lollipops")
	mutColor = flag.String("mut-color", "#ff0000", "color to use for non-synonymous lollipops")

	fontPath = flag.String("f", "", "Path to truetype font to use for drawing (defaults to Arial.ttf)")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] {-Q UNIPROT_DB IDENTIFER | -U UNIPROT_ID | GENE_SYMBOL} [PROTEIN CHANGES ...]\n", os.Args[0])
		fmt.Fprint(os.Stderr, `
Protein ID input:
  GENE_SYMBOL is the official human HGNC gene symbol. This will use the
  UniprotKB API to lookup the UNIPROT_ID.

  You can provide a UniProt ID directly with -U (e.g. "-U P04637" for TP53)

  For more advanced usage, query UniprotKB's database mappings directly using
  a supported identifier with -Q DBNAME. Available DBNAMEs can be found here:
     http://www.uniprot.org/help/programmatic_access#id_mapping_examples

     RefSeq ID        e.g. -Q P_REFSEQ_AC NP_001265252.1
     Entrez GeneID    e.g. -Q P_ENTREZGENEID 4336
     Ensembl ID       e.g. -Q ENSEMBL_ID ENSG00000168314

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
  -legend                 draw a legend for colored regions
  -syn-color="#0000ff"    color to use for synonymous mutation markers
  -mut-color="#ff0000"    color to use for non-synonymous mutation markers
  -hide-axis              do not draw the amino position x-axis
  -show-disordered        draw disordered regions on the backbone
  -show-motifs            draw simple motif regions
  -labels                 draw label text above lollipop markers
  -no-patterns            use solid fill instead of patterns (SVG only)
  -domain-labels=fit      hot to apply domain labels (default="truncated")
                            "fit" = only if fits in space available
                            "off" = do not draw text in the domains

Output options:
  -o=filename.png         set output filename (.png or .svg supported)
  -w=700                  set diagram pixel width (default = automatic fit)
  -dpi=300                set DPI (PNG output only)
`)
	}

	flag.Parse()
	drawing.DefaultSettings.ShowLegend = *showLegend
	drawing.DefaultSettings.ShowLabels = *showLabels
	drawing.DefaultSettings.HideDisordered = !*showDisordered
	drawing.DefaultSettings.HideMotifs = !*showMotifs
	drawing.DefaultSettings.HideAxis = *hideAxis
	drawing.DefaultSettings.SolidFillOnly = *noPatterns
	drawing.DefaultSettings.DomainLabelStyle = *domainLabels
	drawing.DefaultSettings.SynonymousColor = *synColor
	drawing.DefaultSettings.MutationColor = *mutColor
	drawing.DefaultSettings.GraphicWidth = float64(*width)

	if *fontPath == "" {
		err := drawing.LoadDefaultFont()
		if err != nil {
			fmt.Fprintln(os.Stderr, "ERROR: Unable to find Arial.ttf - Which is required for accurate font sizing.")
			fmt.Fprintln(os.Stderr, "       Please use -f=/path/to/arial.ttf or the TrueType (.ttf) font of your choice.")
			// continue in the hopes that SVG rendering will be ok...
			//os.Exit(1)
		}
	} else {
		fname := path.Base(*fontPath)
		fname = strings.TrimSuffix(fname, path.Ext(fname))
		err := drawing.LoadFont(fname, *fontPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	var err error
	varStart := 0
	acc := ""
	geneSymbol := ""
	if *uniprot == "" && flag.NArg() > 0 {
		geneSymbol = flag.Arg(0)
		varStart = 1

		if *queryDB == "GENENAME" {
			fmt.Fprintln(os.Stderr, "HGNC Symbol: ", flag.Arg(0))
			acc, err = data.GetProtID(flag.Arg(0))
		} else {
			fmt.Fprintln(os.Stderr, "Searching for ID: ", flag.Arg(0))
			acc, err = data.GetProtMapping(*queryDB, flag.Arg(0))
		}

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

		if mousetrap.StartedByExplorer() {
			fmt.Fprintln(os.Stderr, `!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!

This is a command-line utility for pipeline processing, you probably don't want
to double-click it! Open your command prompt with 'cmd.exe' and try again.

Press Enter/Ctrl-C to quit.`)
			fmt.Scanln(&acc)
		}
		os.Exit(1)
	}

	var d *data.GraphicResponse = &data.GraphicResponse{}

	length, err := data.GetProtLength(acc)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	d.Length = json.Number(fmt.Sprint(length))

	regions, err := data.GetPfamProteinMatches(acc)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	d.Regions = regions

	motifs, err := data.GetSequenceFeatures(acc)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	d.Motifs = motifs

	if *output == "" {
		*output = geneSymbol + ".svg"
	}

	err = createOutput(*output, d, flag.Args()[varStart:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
