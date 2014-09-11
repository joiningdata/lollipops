lollipops
---------

A simple 'lollipop' mutation diagram generator written in Go. It uses the
[Pfam API](http://pfam.xfam.org/help#tabview=tab9) to retrieve domains and
colors, and the [BioMart API](http://www.biomart.org/) to translate HGNC
Gene Symbols into Uniprot/SwissProt Accession number. If variant changes
are provided, it will also annotate them to the diagram using the
"lollipops" markers that give the tool it's name.

Example
-------

    ./lollipops -w 600 -o tp53.svg TP53 R248Q R273C R249S

Results in the following SVG image:

![TP53 Lollipop diagram with 3 marked mutations](tp53.png?raw=true)