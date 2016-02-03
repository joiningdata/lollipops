//
//    Lollipops diagram generation framework for genetic variations.
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

	// DomainLabelStyle determines how to deal with domain labels that do not fit
	// within the colored domain blocks. Values are "off", "fit" (only labels that
	// fully fit), and "truncated" (default, remove text to fit within).
	DomainLabelStyle string

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

	dpi float64
}

// DefaultSettings contains the "standard" diagram output config and is used by
// the package-level Draw invocations.
var DefaultSettings = Settings{
	ShowLabels:     false,
	HideDisordered: false,
	HideMotifs:     false,
	HideAxis:       false,
	SolidFillOnly:  false,

	DomainLabelStyle: "truncated",

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
