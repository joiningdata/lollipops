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

    ./lollipops -w=700 TP53 R248Q R273C R249S

Results in the following SVG image:

![TP53 Lollipop diagram with 3 marked mutations](tp53.png?raw=true)

Usage
-----

Usage: ``lollipops [options] GENE_SYMBOL [PROTEIN CHANGES ...]``

Where ``GENE_SYMBOL`` is the official HGNC symbol and ``PROTEIN CHANGES``
is a list of amino acid changes of the format "(amino-code)(position)..."
Amino-code can be either the 1- or 3-character code for the amino acid.
Only the first position in each change is used for plotting even if the
change contains a range. All characters after the position are ignored.

    -o=out.svg         SVG output filename (default GENE_SYMBOL.svg)
    -labels            draw labels for each mutation
    -hide-axis         do not draw the aa position axis
    -hide-disordered   do not draw disordered regions
    -hide-motifs       do not draw motifs
    -w=700             SVG output width (default=automatic)

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