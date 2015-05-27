package drawing

import "flag"

var (
	showLabels     = flag.Bool("labels", false, "draw mutation labels above lollipops")
	hideDisordered = flag.Bool("hide-disordered", false, "do not draw disordered regions")
	hideMotifs     = flag.Bool("hide-motifs", false, "do not draw motifs")
	hideAxis       = flag.Bool("hide-axis", false, "do not draw the aa position axis")
	forPDF         = flag.Bool("for-pdf", false, "use solid fill instead of patterns for PDF output")

	synColor = flag.String("syn-color", "#0000ff", "color to use for synonymous lollipops")
	mutColor = flag.String("mut-color", "#ff0000", "color to use for non-synonymous lollipops")
)

const (
	LollipopRadius = 4
	LollipopHeight = 28
	BackboneHeight = 14
	MotifHeight    = 18
	DomainHeight   = 24
	Padding        = 15
	AxisPadding    = 10
	AxisHeight     = 15
	TextPadding    = 5
	GraphicHeight  = DomainHeight + Padding*2
	//GraphicWidth   = 740
)
