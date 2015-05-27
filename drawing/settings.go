package drawing

// Settings contains all the configurable options for lollipop diagram generation.
type Settings struct {
	// ShowLabels adds mutation label text above lollipops markers.
	ShowLabels bool
	// HideDisordered hides disordered regions on the backbone even if motifs are shown.
	HideDisordered bool
	// HideMotifs hides motifs in the output image.
	HideMotifs bool
	// HideAxis hides the amino acid position axis in the output image.
	HideAxis bool

	// SolidFillOnly ensures no patterns are used in output files.
	SolidFillOnly bool

	// SynonymousColor is the #RRGGBB color to use for synonymous mutations.
	SynonymousColor string
	// MutationColor is the #RRGGBB color to use for non-synonymous mutations.
	MutationColor string

	// LollipopRadius is the size of the marker at the top of the "stick".
	LollipopRadius float64
	// LollipopHeight is the length of the "stick" connecting the backbone to the marker.
	LollipopHeight float64
	// BackboneHeight is the thickness of the amino acid backbone.
	BackboneHeight float64
	// MotifHeight is the thickness of a motif region.
	MotifHeight float64
	// DomainHeight is the thickness of a domain region.
	DomainHeight float64
	// Padding is the amount of whitespace added to each side of the image.
	Padding float64
	// AxisPadding is the amount of whitespace added between the axis and backbone.
	AxisPadding float64
	// AxisHeight is the height of the axis tick lines.
	AxisHeight float64
	// TextPadding is the amount of whitespace between the axis line and text.
	TextPadding float64

	// GraphicWidth is the width of the image, if <=0 then it will be automatically
	// determined based on the image contents.
	GraphicWidth float64

	// GraphicHeight is automatically determined based on configured options.
	GraphicHeight float64
}

// DefaultSettings contains the "standard" diagram output config and is used by
// the package-level Draw invocations.
var DefaultSettings = Settings{
	ShowLabels:     false,
	HideDisordered: false,
	HideMotifs:     false,
	HideAxis:       false,
	SolidFillOnly:  false,

	SynonymousColor: "#0000ff",
	MutationColor:   "#ff0000",

	LollipopRadius: 4,
	LollipopHeight: 28,
	BackboneHeight: 14,
	MotifHeight:    18,
	DomainHeight:   24,
	Padding:        15,
	AxisPadding:    10,
	AxisHeight:     15,
	TextPadding:    5,
}
