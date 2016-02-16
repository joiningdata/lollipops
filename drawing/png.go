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
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"log"

	"github.com/golang/freetype/truetype"
	"github.com/pbnjay/lollipops/data"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

func DrawPNG(w io.Writer, dpi float64, changelist []string, g *data.PfamGraphicResponse) {
	DefaultSettings.dpi = 0
	DefaultSettings.DrawPNG(w, dpi, changelist, g)
}

// DrawPNG writes PNG image to w, with the provided changes in changelist and
// Pfam domain/region information in g. If GraphicWidth=0, then AutoWidth is called
// to determine the best diagram width to fit all labels.
func (s *Settings) DrawPNG(w io.Writer, dpi float64, changelist []string, g *data.PfamGraphicResponse) {
	if s.dpi == 0 {
		dpiScale := dpi / 72.0
		s.LollipopRadius *= dpiScale
		s.LollipopHeight *= dpiScale
		s.BackboneHeight *= dpiScale
		s.MotifHeight *= dpiScale
		s.DomainHeight *= dpiScale
		s.Padding *= dpiScale
		s.AxisPadding *= dpiScale
		s.AxisHeight *= dpiScale
		s.TextPadding *= dpiScale
		s.dpi = dpi
	}
	d := s.prepare(changelist, g)
	d.png(w)
}

func (s *diagram) png(w io.Writer) {
	aaLen, _ := s.g.Length.Int64()
	scale := (s.GraphicWidth - s.Padding*2) / float64(aaLen)
	aaSpace := int((20 * s.dpi / 72.0) / scale)

	img := image.NewRGBA(image.Rect(0, 0, int(s.GraphicWidth), int(s.GraphicHeight)))
	drawRectWH(img, 0, 0, s.GraphicWidth, s.GraphicHeight, color.White)

	blackFontDrawer := &font.Drawer{
		Dst: img,
		Src: &image.Uniform{color.Black},
		Face: truetype.NewFace(theFont, &truetype.Options{
			Size:    float64(10.0),
			DPI:     float64(DefaultSettings.dpi),
			Hinting: font.HintingFull,
		}),
	}
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

		c := color.RGBA{0xBA, 0xBD, 0xB6, 0xFF}
		thickvline(img, int(pop.x-s.dpi/144), int(pop.y), int(popbot), 2*s.dpi/72.0, c)
		drawCircle(img, int(pop.x+s.dpi/144), int(pop.y), int(pop.r), colorFromHex(pop.Col))

		if s.ShowLabels {
			chg := pop.label
			if pop.Cnt > 1 {
				chg = fmt.Sprintf("%s (%d)", chg, pop.Cnt)
			}

			// FIXME: rotate label to match SVG output
			wf := blackFontDrawer.MeasureString(chg)
			blackFontDrawer.Dot = fixed.Point26_6{
				X: fixed.I(int(pop.x)) - wf/2,
				Y: fixed.I(int(pop.y - (pop.r * 1.5))),
			}
			blackFontDrawer.DrawString(chg)
		}
	}

	// draw the backbone
	drawRectWH(img, s.Padding, startY+(s.DomainHeight-s.BackboneHeight)/2, s.GraphicWidth-(s.Padding*2),
		s.BackboneHeight, color.RGBA{0xBA, 0xBD, 0xB6, 0xFF})

	if !s.HideMotifs {
		disFill := color.RGBA{0, 0, 0, 38} // 15% opacity

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

			if r.Type == "disorder" {
				// draw disordered regions with a understated diagonal hatch pattern
				drawRectWH(img, s.Padding+sstart, startY+(s.DomainHeight-s.BackboneHeight)/2,
					swidth, s.BackboneHeight, disFill)
			} else {
				drawRectWHShadow(img, s.Padding+sstart, startY+(s.DomainHeight-s.MotifHeight)/2,
					swidth, s.MotifHeight, colorFromHex(BlendColorStrings(r.Color, "#FFFFFF")),
					2*s.dpi/72.0)
			}
		}
	}

	whiteFontDrawer := &font.Drawer{
		Dst: img,
		Src: &image.Uniform{color.White},
		Face: truetype.NewFace(theFont, &truetype.Options{
			Size:    float64(12.0),
			DPI:     float64(DefaultSettings.dpi),
			Hinting: font.HintingFull,
		}),
	}

	// get font height in px
	bounds, _, ok := whiteFontDrawer.Face.GlyphBounds('M')
	if !ok {
		log.Fatalf("unable to determine font bounds!")
	}
	// add 2px line spacing
	fontH := bounds.Max.Sub(bounds.Min).Y + fixed.I(2)

	// draw the curated domains
	for ri, r := range s.g.Regions {
		sstart, _ := r.Start.Float64()
		swidth, _ := r.End.Float64()

		sstart *= scale
		swidth = (swidth * scale) - sstart

		drawRectWHShadow(img, s.Padding+sstart, startY, swidth, s.DomainHeight, colorFromHex(r.Color), 2*s.dpi/72.0)

		if swidth > 10 && s.domainLabels[ri] != "" {
			// center text at x
			wf := whiteFontDrawer.MeasureString(s.domainLabels[ri])
			whiteFontDrawer.Dot = fixed.Point26_6{
				X: fixed.I(int(s.Padding+sstart)) + (fixed.I(int(swidth))-wf)/2,
				Y: fixed.I(int(startY+s.DomainHeight/2)) + fontH/2,
			}
			whiteFontDrawer.DrawString(s.domainLabels[ri])
		}
	}

	if !s.HideAxis {
		startY += s.DomainHeight + s.AxisPadding
		thickhline(img, int(s.Padding), int(s.GraphicWidth-s.Padding+s.dpi/36.0), int(startY), s.dpi/72.0, color.Gray{0xAA})
		thickvline(img, int(s.Padding), int(startY), int(startY+(s.AxisHeight/3)), s.dpi/72.0, color.Gray{0xAA})

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
			thickvline(img, int(x), int(startY), int(startY+(s.AxisHeight/3)), s.dpi/72.0, color.Gray{0xAA})

			// center text at x
			spos := fmt.Sprint(t.Pos)

			wf := blackFontDrawer.MeasureString(spos)
			blackFontDrawer.Dot = fixed.Point26_6{
				X: fixed.I(int(x)) - wf/2,
				Y: fixed.I(int(startY + s.AxisHeight)),
			}
			blackFontDrawer.DrawString(spos)
		}
	}

	png.Encode(w, img)
}

///////////

func colorFromHex(h string) color.RGBA {
	c := color.RGBA{0, 0, 0, 0xFF}
	fmt.Sscanf(h, "#%02X%02X%02X", &c.R, &c.G, &c.B)
	return c
}

func thickhline(img draw.Image, x0, x1, y int, dpiScale float64, clr color.Color) {
	for y0 := float64(y); y0 <= float64(y)+dpiScale; y0++ {
		hline(img, x0, x1, int(y0), clr)
	}
}

func thickvline(img draw.Image, x, y0, y1 int, dpiScale float64, clr color.Color) {
	for x0 := float64(x); x0 <= float64(x)+dpiScale; x0++ {
		vline(img, int(x0), y0, y1, clr)
	}
}

func hline(img draw.Image, x0, x1, y int, clr color.Color) {
	for x := int(x0); x <= int(x1); x++ {
		img.Set(x, int(y), clr)
	}
}

func vline(img draw.Image, x, y0, y1 int, clr color.Color) {
	for y := int(y0); y <= int(y1); y++ {
		img.Set(int(x), y, clr)
	}
}

func drawRectWH(img draw.Image, x0, y0, w, h float64, clr color.Color) {
	draw.Draw(img, image.Rect(int(x0), int(y0), int(x0+w), int(y0+h)),
		&image.Uniform{clr}, image.ZP, draw.Over)
}

func drawRectWHShadow(img draw.Image, x0, y0, w, h float64, clr color.Color, shadowOffs float64) {
	// approx 10% opacity
	src := &image.Uniform{color.RGBA{0, 0, 0, 1 + uint8(75/shadowOffs)}}
	for i := shadowOffs; i > 0; i-- {
		r := image.Rect(int(x0+i), int(y0+i), int(x0+i+w), int(y0+i+h))
		draw.Draw(img, r, src, image.ZP, draw.Over)
	}
	r := image.Rect(int(x0), int(y0), int(x0+w), int(y0+h))
	draw.Draw(img, r, &image.Uniform{clr}, image.ZP, draw.Over)
}

// http://en.wikipedia.org/wiki/Midpoint_circle_algorithm
func drawCircle(img draw.Image, x0, y0, radius int, clr color.Color) {
	f := 1.0 - radius
	dx, dy := 1, -2*radius
	x, y := 0, radius

	hline(img, x0-radius, x0+radius, y0, clr)
	vline(img, x0, y0-radius, y0+radius, clr)

	for x < y {
		if f >= 0 {
			y--
			dy += 2
			f += dy
		}
		x++
		dx += 2
		f += dx
		hline(img, x0-x, x0+x, y0+y, clr)
		hline(img, x0-x, x0+x, y0-y, clr)
		hline(img, x0-y, x0+y, y0+x, clr)
		hline(img, x0-y, x0+y, y0-x, clr)
	}
}
