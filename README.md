#lollipops

A simple 'lollipop' mutation diagram generator that tries to make things
simple and easy by automating as much as possible. It uses the
[Pfam API](http://pfam.xfam.org/help#tabview=tab9) to retrieve domains and
colors, and the [UniProt REST API](http://www.uniprot.org/uploadlists/) to
translate HGNC Gene Symbols into Uniprot/SwissProt Accession number. If
variant changes are provided, it will also annotate them to the diagram
using the "lollipops" markers that give the tool it's name.

##Example

Basic usage is just the gene symbol (ex: ``TP53``) and a list of
mutations (ex: ``R273C R175H T125 R248Q``)

    ./lollipops TP53 R273C R175H T125 R248Q

![TP53 Lollipop diagram with 4 marked mutations](tp53.png?raw=true)

More advanced usage allows for per-mutation color (e.x. sample type) and
size specification (i.e. denoting number of samples), along with text
labels and more:

		./lollipops -labels TP53 R248Q#7f3333@131 R273C R175H T125@5

![TP53 Lollipop diagram with 5 customized mutations](tp53_more.png?raw=true)

##Usage

Usage: ``lollipops [options] {-U UNIPROT_ID | GENE_SYMBOL} [PROTEIN CHANGES ...]``

Where **GENE_SYMBOL** is the official human HGNC gene symbol. This will use the
BioMart API to lookup the **UNIPROT_ID**. To skip the lookup or use other species,
specify the UniProt ID with -U (e.g. ``-U P04637`` for TP53)

####Protein changes

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

**N.B.** Color must come before count in tags.

####Diagram generation options

```
  -syn-color="#0000ff"    color to use for synonymous mutation markers
  -mut-color="#ff0000"    color to use for non-synonymous mutation markers
  -hide-axis              do not draw the amino position x-axis
  -hide-disordered        do not draw disordered regions on the backbone
  -hide-motifs            do not draw simple motif regions
  -labels                 draw label text above lollipop markers
  -no-patterns            use solid fill instead of patterns (SVG only)
```

####Output options

```
  -o=filename.png         set output filename (.png or .svg supported)
  -w=700                  set diagram pixel width (default = automatic fit)
  -dpi=300                set DPI (PNG output only)
```


##Installation

Head over to the [Releases](https://github.com/pbnjay/lollipops/releases) to
download the latest version for your system in a simple command-line executable.

If you already have Go installed and want the bleeding edge, just
``go get -u github.com/pbnjay/lollipops`` to download the latest version.

##Embedding

As of v0.97, lollipops is now easy to embed in other Go applications. The following code prints an SVG for TP53 and some mutations to standard output:

```go
package main

import (
    "os"

    "github.com/pbnjay/lollipops/data"
    "github.com/pbnjay/lollipops/drawing"
)

func main() {
    uniprot_id := "P04637"
    mutations := []string{"R273C", "R175H", "T125@5", "R248Q#7f3333@131"}

    p53_domains, err := data.GetPfamGraphicData(uniprot_id)
    if err != nil {
        panic(err)
    }

    drawing.DrawSVG(os.Stdout, mutations, p53_domains)
}

```

