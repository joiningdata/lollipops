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

import (
	"fmt"
	"io"

	"github.com/pbnjay/lollipops/data"
)

const svgHeader = `<?xml version='1.0'?>
<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="%f" height="%f">
<defs>
  <filter id="ds" x="0" y="0">
    <feOffset in="SourceAlpha" dx="2" dy="2" />
    <feComponentTransfer><feFuncA type="linear" slope="0.2"/></feComponentTransfer>
    <feGaussianBlur result="blurOut" stdDeviation="1" />
    <feBlend in="SourceGraphic" in2="blurOut" mode="normal" />
  </filter>
  <pattern id="disordered-hatch" patternUnits="userSpaceOnUse" width="4" height="4">
    <path d="M-1,1 l2,-2 M0,4 l4,-4 M3,5 l2,-2" stroke="#000000" opacity="0.3" />
  </pattern>
</defs>
`
const svgFooter = `</svg>`

func DrawSVG(w io.Writer, changelist []string, g *data.GraphicResponse) {
	d := DefaultSettings.prepare(changelist, g)
	d.svg(w)
}

// DrawSVG writes the SVG XML document to w, with the provided changes in changelist
// and domain/region information in g. If GraphicWidth=0, the AutoWidth is called
// to determine the best diagram width to fit all labels.
func (s *Settings) DrawSVG(w io.Writer, changelist []string, g *data.GraphicResponse) {
	d := s.prepare(changelist, g)
	d.svg(w)
}

func (s *diagram) svg(w io.Writer) {
	aaLen, _ := s.g.Length.Int64()
	scale := (s.GraphicWidth - s.Padding*2) / float64(aaLen)
	aaSpace := int(20 / scale)
	fontSpec := ""
	if FontName != "" {
		fontSpec = "font-family:" + FontName + ";"
	}

	fmt.Fprintf(w, svgHeader, s.GraphicWidth, s.GraphicHeight)

	//////

	startY := s.startY
	poptop := startY + s.LollipopRadius
	popbot := poptop + s.LollipopHeight

	firstLollipop := true
	for _, pop := range s.ticks {
		if !pop.isLollipop {
			continue
		}
		if firstLollipop {
			firstLollipop = false
			startY = popbot - (s.DomainHeight-s.BackboneHeight)/2
		}

		fmt.Fprintf(w, `<line x1="%f" x2="%f" y1="%f" y2="%f" stroke="#BABDB6" stroke-width="2"/>`, pop.x, pop.x, pop.y, popbot)
		fmt.Fprintf(w, `<a xlink:title="%s"><circle cx="%f" cy="%f" r="%f" fill="%s" /></a>`,
			pop.label, pop.x, pop.y, pop.r, pop.Col)

		if s.ShowLabels {
			fmt.Fprintf(w, `<g transform="translate(%f,%f) rotate(-30)">`,
				pop.x, pop.y)
			chg := pop.label
			if pop.Cnt > 1 {
				chg = fmt.Sprintf("%s (%d)", chg, pop.Cnt)
			}
			fmt.Fprintf(w, `<text style="font-size:10px;%sfill:#555;" text-anchor="middle" x="0" y="%f">%s</text></g>`,
				fontSpec, (pop.r * -1.5), chg)
		}
	}

	// draw the backbone
	fmt.Fprintf(w, `<a xlink:title="%s, %s (%daa)"><rect fill="#BABDB6" x="%f" y="%f" width="%f" height="%f"/></a>`,
		s.g.Metadata.Identifier, s.g.Metadata.Description, aaLen,
		s.Padding, startY+(s.DomainHeight-s.BackboneHeight)/2, s.GraphicWidth-(s.Padding*2), s.BackboneHeight)

	disFill := "url(#disordered-hatch)"
	if s.SolidFillOnly {
		disFill = `#000;" opacity="0.15`
	}
	if !s.HideMotifs {
		// draw transmembrane, signal peptide, coiled-coil, etc motifs
		for _, r := range s.g.Motifs {
			if r.Type == "pfamb" {
				continue
			}
			if r.Type == "disorder" && s.HideDisordered {
				continue
			}
			sstart, _ := r.Start.Float64()
			swidth, _ := r.End.Float64()

			sstart *= scale
			swidth = (swidth * scale) - sstart

			fmt.Fprintf(w, `<a xlink:title="%s">`, r.Type)
			if r.Type == "disorder" {
				// draw disordered regions with a understated diagonal hatch pattern
				fmt.Fprintf(w, `<rect fill="%s" x="%f" y="%f" width="%f" height="%f"/>`, disFill,
					s.Padding+sstart, startY+(s.DomainHeight-s.BackboneHeight)/2, swidth, s.BackboneHeight)
			} else {
				fmt.Fprintf(w, `<rect fill="%s" x="%f" y="%f" width="%f" height="%f" filter="url(#ds)"/>`, BlendColorStrings(r.Color, "#FFFFFF"),
					s.Padding+sstart, startY+(s.DomainHeight-s.MotifHeight)/2, swidth, s.MotifHeight)
			}
			fmt.Fprintln(w, `</a>`)
		}
	}

	// draw the curated domains
	for ri, r := range s.g.Regions {
		sstart, _ := r.Start.Float64()
		swidth, _ := r.End.Float64()

		sstart *= scale
		swidth = (swidth * scale) - sstart

		fmt.Fprintf(w, `<g transform="translate(%f,%f)"><a xlink:href="%s" xlink:title="%s">`, s.Padding+sstart, startY, r.Link, r.Metadata.Description)
		fmt.Fprintf(w, `<rect fill="%s" x="0" y="0" width="%f" height="%f" filter="url(#ds)"/>`, r.Color, swidth, s.DomainHeight)
		if swidth > 10 && s.domainLabels[ri] != "" {
			fmt.Fprintf(w, `<text style="font-size:12px;%sfill:#ffffff;" text-anchor="middle" x="%f" y="%f">%s</text>`,
				fontSpec, swidth/2.0, 4+s.DomainHeight/2, s.domainLabels[ri])
		}
		fmt.Fprintln(w, `</a></g>`)
	}

	if !s.HideAxis {
		startY += s.DomainHeight + s.AxisPadding
		fmt.Fprintln(w, `<g class="axis">`)
		fmt.Fprintf(w, `<line x1="%f" x2="%f" y1="%f" y2="%f" stroke="#AAAAAA" />`, s.Padding, s.GraphicWidth-s.Padding, startY, startY)
		fmt.Fprintf(w, `<line x1="%f" x2="%f" y1="%f" y2="%f" stroke="#AAAAAA" />`, s.Padding, s.Padding, startY, startY+(s.AxisHeight/3))

		lastDrawn := 0
		for i, t := range s.ticks {
			if lastDrawn > 0 && (t.Pos-lastDrawn) < aaSpace {
				continue
			}
			j := s.ticks.NextBetter(i, aaSpace)
			if i != j {
				continue
			}
			lastDrawn = t.Pos
			x := s.Padding + (float64(t.Pos) * scale)
			fmt.Fprintf(w, `<line x1="%f" x2="%f" y1="%f" y2="%f" stroke="#AAAAAA" />`, x, x, startY, startY+(s.AxisHeight/3))
			fmt.Fprintf(w, `<text style="font-size:10px;%sfill:#000000;" text-anchor="middle" x="%f" y="%f">%d</text>`,
				fontSpec, x, startY+s.AxisHeight, t.Pos)
		}

		fmt.Fprintln(w, "</g>")
		startY += s.AxisHeight
	}

	for key, color := range s.legendInfo {
		startY += 14.0
		if key == data.MotifNames["disorder"] {
			color = disFill
		}
		fmt.Fprintf(w, `<rect fill="%s" x="4" y="%f" width="12" height="12" filter="url(#ds)"/>`, color, startY)
		fmt.Fprintf(w, `<text style="font-size:12px;%sfill:#000000;" text-anchor="start" x="20" y="%f">%s</text>`,
			fontSpec, startY+12, key) // 12=font height-baseline
	}

	fmt.Fprintln(w, svgFooter)
}
