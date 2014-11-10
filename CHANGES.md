Added new field to Tick struct for color change.  Fixed regex so it can actually capture three letter or one letter amino acid annotations.  Turns lollipop red for non-synonymous, blue for synonymous.

For synonymous mutations the format is reference amino acid then amino acid number.  For non-synonymous it is reference amino acid then amino acid number then alternate amino acid.  Like so:
	non-syn: E234W
	syn: E234
