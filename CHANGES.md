Added new field to Tick struct for color change.  Fixed regex so it can actually capture three letter or one letter amino acid annotations.  Turns lollipop red for non-synonymous, blue for synonymous.

For synonymous mutations the format is reference amino acid then amino acid number.  For non-synonymous it is reference amino acid then amino acid number then alternate amino acid.  Like so:
	non-syn: TP53 E234W
	syn: TP53 E234

command used to generate png:
lollipops TP53	E343Q R342Q F338 R335C R283H R283C R282W R248Q G245S C242 N235S P223H P222 P222L V216 R213 L206 L194 D186E S185 S185N R158 R156H Y107H
