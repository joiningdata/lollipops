lollipops
---------

A simple 'lollipop' mutation diagram generator that tries to make things
simple and easy by automating as much as possible. It uses the
[Pfam API](http://pfam.xfam.org/help#tabview=tab9) to retrieve domains and
colors, and the [BioMart API](http://www.biomart.org/) to translate HGNC
Gene Symbols into Uniprot/SwissProt Accession number. If variant changes
are provided, it will also annotate them to the diagram using the
"lollipops" markers that give the tool it's name.

Example
-------

Basic usage is just the gene symbol (ex: ``TP53``) and a list of
mutations (ex: ``R273C R175H T125 R248Q``)

    ./lollipops TP53 R273C R175H T125 R248Q

![TP53 Lollipop diagram with 4 marked mutations](tp53.png?raw=true)

More advanced usage allows for per-mutation color (e.x. sample type) and
size specification (i.e. denoting number of samples), along with text
labels and more:

		./lollipops -labels TP53 R248Q#7f3333@131 R273C R175H T125@5

![TP53 Lollipop diagram with 5 customized mutations](tp53_more.png?raw=true)

Usage
-----

Usage: ``lollipops [options] GENE_SYMBOL [PROTEIN CHANGES ...]``

Where ``GENE_SYMBOL`` is the official HGNC symbol and ``PROTEIN CHANGES``
is a list of amino acid changes of the format "(amino-code)(position)..."
Amino-code can be either the 1- or 3-character code for the amino acid.
Only the first position in each change is used for plotting even if the
change contains a range. All characters after the position are ignored.
Protein changes may also be appended with a hex color code (seen in
example above) to alter the lollipop color for each specific mutation.

    -o=out.svg         SVG output filename (default GENE_SYMBOL.svg)
    -labels            draw labels for each mutation
    -hide-axis         do not draw the aa position axis
    -hide-disordered   do not draw disordered regions
    -hide-motifs       do not draw motifs
    -w=700             SVG output width (default=automatic)
    -mut-color=#ff0000 color to use for non-synonymous lollipops
    -syn-color=#0000ff color to use for synonymous lollipops

If you are working with non-human data, or know the Uniprot Accession
already, You can specify it with `-U UNIPROTID` instead of GENE_SYMBOL,
for example the following mouse query works for gene `Mobp`:

    ./lollipops -U Q9D2P8

Installation
------------

Head over to the [Releases](https://github.com/pbnjay/lollipops/releases) to
download the latest version for your system in a simple command-line executable.

If you already have Go installed and want the bleeding edge, just
``go get -u github.com/pbnjay/lollipops`` to download the latest version.
