package main

import (
	"flag"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

var (
	showLabels     = flag.Bool("labels", false, "draw mutation labels above lollipops")
	hideDisordered = flag.Bool("hide-disordered", false, "do not draw disordered regions")
	hideMotifs     = flag.Bool("hide-motifs", false, "do not draw motifs")
	hideAxis       = flag.Bool("hide-axis", false, "do not draw the aa position axis")
	forPDF         = flag.Bool("for-pdf", false, "use solid fill instead of patterns for PDF output")
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

const svgHeader = `<?xml version='1.0'?>
<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="%d" height="%d">
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

type Tick struct {
	Pos int
	Pri int
	Col string
}

type TickSlice []Tick

func (t TickSlice) NextBetter(i, maxDist int) int {
	for j := i; j < len(t); j++ {
		if (t[j].Pos - t[i].Pos) > maxDist {
			return i
		}
		if t[j].Pri > t[i].Pri {
			return j
		}
	}
	return i
}

// implement sort interface
func (t TickSlice) Len() int      { return len(t) }
func (t TickSlice) Swap(i, j int) { t[i], t[j] = t[j], t[i] }
func (t TickSlice) Less(i, j int) bool {
	if t[i].Pos == t[j].Pos {
		// sort high-priority first if same
		return t[i].Pri > t[j].Pri
	}
	return t[i].Pos < t[j].Pos
}

var stripChangePos = regexp.MustCompile("(^|[A-Za-z]*)([0-9]+)([A-Za-z]*)")

// BlendColorStrings blends two CSS #RRGGBB colors together with a straight average.
func BlendColorStrings(a, b string) string {
	var r1, g1, b1, r2, g2, b2 int
	fmt.Sscanf(strings.ToUpper(a), "#%02X%02X%02X", &r1, &g1, &b1)
	fmt.Sscanf(strings.ToUpper(b), "#%02X%02X%02X", &r2, &g2, &b2)
	return fmt.Sprintf("#%02X%02X%02X", (r1+r2)/2, (g1+g2)/2, (b1+b2)/2)
}

// AutoWidth automatically determines the best width to use to fit all
// available domain names into the plot.
func AutoWidth(g *PfamGraphicResponse) int {
	aaLen, _ := g.Length.Float64()
	w := 400.0

	for _, r := range g.Regions {
		sstart, _ := r.Start.Float64()
		send, _ := r.End.Float64()

		aaPart := (send - sstart) / aaLen
		minTextWidth := MeasureFont(r.Text, 12) + (TextPadding * 2) + 1

		ww := (float64(minTextWidth) / aaPart)
		if ww > w {
			w = ww
		}
	}
	return int(w + (Padding * 2))
}

// DrawSVG writes the SVG XML document to w, with the provided changes in changelist
// and Pfam domain/region information in g. If GraphicWidth=0, the AutoWidth is called
// to determine the best diagram width to fit all labels.
func DrawSVG(w io.Writer, GraphicWidth int, changelist []string, g *PfamGraphicResponse) {
	if GraphicWidth == 0 {
		GraphicWidth = AutoWidth(g)
	}
	aaLen, _ := g.Length.Int64()
	scale := float64(GraphicWidth-Padding*2) / float64(aaLen)
	popSpace := int(float64(LollipopRadius+2) / scale)
	aaSpace := int(20 / scale)
	startY := Padding
	if *showLabels {
		startY += Padding // add some room for labels
	}

	pops := TickSlice{}
	col := "#0000ff"
	ht := GraphicHeight
	if len(changelist) > 0 {
		// parse changelist and check if lollipops need staggered
		for i, chg := range changelist {
			cpos := stripChangePos.FindStringSubmatch(chg)
			spos := 0
			if cpos[3] != "" {
				col = "FF5555"
			} else if cpos[3] == "" {
				col = "#0000ff"
			}
			if strings.Contains(chg, "#") {
				col = "#" + strings.SplitN(chg, "#", 2)[1]
			}
			fmt.Sscanf(cpos[2], "%d", &spos)
			pops = append(pops, Tick{spos, -i, col})
		}
		sort.Sort(pops)
		maxStaggered := LollipopRadius + LollipopHeight
		for pi, pop := range pops {
			h := LollipopRadius + LollipopHeight
			for pj := pi + 1; pj < len(pops); pj++ {
				if pops[pj].Pos-pop.Pos > popSpace {
					break
				}
				h += LollipopRadius * 3
			}
			if h > maxStaggered {
				maxStaggered = h
			}
		}
		ht += maxStaggered
		startY += maxStaggered - (LollipopRadius + LollipopHeight)
	}
	if !*hideAxis {
		ht += AxisPadding + AxisHeight
	}

	ticks := []Tick{
		Tick{0, 0, col},           // start isn't very important (0 is implied) // wrote in new field - Jim H.
		Tick{int(aaLen), 99, col}, // always draw the length in the axis // wrote in new field - Jim H.
	}

	fmt.Fprintf(w, svgHeader, GraphicWidth, ht)

	if len(pops) > 0 {
		poptop := startY + LollipopRadius
		popbot := poptop + LollipopHeight
		startY = popbot - (DomainHeight-BackboneHeight)/2

		// draw lollipops
		for pi, pop := range pops {
			ticks = append(ticks, Tick{pop.Pos, 10, col})
			spos := Padding + (float64(pop.Pos) * scale)

			mytop := poptop
			for pj := pi + 1; pj < len(pops); pj++ {
				if pops[pj].Pos-pop.Pos > popSpace {
					break
				}
				mytop -= LollipopRadius * 3
			}
			fmt.Fprintf(w, `<line x1="%f" x2="%f" y1="%d" y2="%d" stroke="#BABDB6" stroke-width="2"/>`, spos, spos, mytop, popbot)
			fmt.Fprintf(w, `<a xlink:title="%s"><circle cx="%f" cy="%d" r="%d" fill="%s" /></a>`,
				changelist[-pop.Pri], spos, mytop, LollipopRadius, pop.Col)

			if *showLabels {
				fmt.Fprintf(w, `<g transform="translate(%f,%d) rotate(-30)">`,
					spos, mytop)
				fmt.Fprintf(w, `<text style="font-size:10px;font-family:sans-serif;fill:#555;" text-anchor="middle" x="0" y="%f">%s</text></g>`,
					(LollipopRadius * -1.5), changelist[-pop.Pri])
			}
		}
	}

	// draw the backbone
	fmt.Fprintf(w, `<a xlink:title="%s, %s (%daa)"><rect fill="#BABDB6" x="%d" y="%d" width="%d" height="%d"/></a>`,
		g.Metadata.Identifier, g.Metadata.Description, aaLen,
		Padding, startY+(DomainHeight-BackboneHeight)/2, GraphicWidth-(Padding*2), BackboneHeight)

	disFill := "url(#disordered-hatch)"
	if *forPDF {
		disFill = `#000;" opacity="0.15`
	}
	if !*hideMotifs {
		// draw transmembrane, signal peptide, coiled-coil, etc motifs
		for _, r := range g.Motifs {
			if r.Type == "pfamb" {
				continue
			}
			if r.Type == "disorder" && *hideDisordered {
				continue
			}
			sstart, _ := r.Start.Float64()
			swidth, _ := r.End.Float64()

			sstart *= scale
			swidth = (swidth * scale) - sstart

			fmt.Fprintf(w, `<a xlink:title="%s">`, r.Type)
			if r.Type == "disorder" {
				// draw disordered regions with a understated diagonal hatch pattern
				fmt.Fprintf(w, `<rect fill="%s" x="%f" y="%d" width="%f" height="%d"/>`, disFill,
					Padding+sstart, startY+(DomainHeight-BackboneHeight)/2, swidth, BackboneHeight)
			} else {
				fmt.Fprintf(w, `<rect fill="%s" x="%f" y="%d" width="%f" height="%d" filter="url(#ds)"/>`, BlendColorStrings(r.Color, "#FFFFFF"),
					Padding+sstart, startY+(DomainHeight-MotifHeight)/2, swidth, MotifHeight)

				tstart, _ := r.Start.Int64()
				tend, _ := r.End.Int64()
				ticks = append(ticks, Tick{int(tstart), 1, col})
				ticks = append(ticks, Tick{int(tend), 1, col})
			}
			fmt.Fprintln(w, `</a>`)
		}
	}

	// draw the curated domains
	for _, r := range g.Regions {
		sstart, _ := r.Start.Float64()
		swidth, _ := r.End.Float64()

		ticks = append(ticks, Tick{int(sstart), 5, col})
		ticks = append(ticks, Tick{int(swidth), 5, col})

		sstart *= scale
		swidth = (swidth * scale) - sstart

		fmt.Fprintf(w, `<g transform="translate(%f,%d)"><a xlink:href="%s" xlink:title="%s">`, Padding+sstart, startY, "http://pfam.xfam.org"+r.Link, r.Metadata.Description)
		fmt.Fprintf(w, `<rect fill="%s" x="0" y="0" width="%f" height="%d" filter="url(#ds)"/>`, r.Color, swidth, DomainHeight)
		if swidth > 10 {
			if len(r.Metadata.Description) > 1 && float64(MeasureFont(r.Metadata.Description, 12)) < (swidth-TextPadding) {
				// we can fit the full description! nice!
				fmt.Fprintf(w, `<text style="font-size:12px;font-family:sans-serif;fill:#ffffff;" text-anchor="middle" x="%f" y="%d">%s</text>`, swidth/2.0, 4+DomainHeight/2, r.Metadata.Description)
			} else if float64(MeasureFont(r.Text, 12)) < (swidth - TextPadding) {
				fmt.Fprintf(w, `<text style="font-size:12px;font-family:sans-serif;fill:#ffffff;" text-anchor="middle" x="%f" y="%d">%s</text>`, swidth/2.0, 4+DomainHeight/2, r.Text)
			} else {
				didOutput := false
				if strings.IndexFunc(r.Text, unicode.IsPunct) != -1 {

					// if the label is too long, we assume the most
					// informative word is the last one, but if that
					// still won't fit we'll move up
					//
					// Example: TP53 has P53_TAD and P53_tetramer
					// domains but boxes aren't quite large enough.
					// Showing "P53..." isn't very helpful.

					parts := strings.FieldsFunc(r.Text, unicode.IsPunct)
					pre := ".."
					post := ""
					for i := len(parts) - 1; i >= 0; i-- {
						if i == 0 {
							pre = ""
						}
						if float64(MeasureFont(pre+parts[i]+post, 12)) < (swidth - TextPadding) {
							fmt.Fprintf(w, `<text style="font-size:12px;font-family:sans-serif;fill:#ffffff;" text-anchor="middle" x="%f" y="%d">%s</text>`, swidth/2.0, 4+DomainHeight/2, pre+parts[i]+post)
							didOutput = true
							break
						}
						post = ".."
					}
				}

				if !didOutput && swidth > 40 {
					sub := r.Text
					for mx := len(r.Text) - 2; mx > 0; mx-- {
						sub = strings.TrimFunc(r.Text[:mx], unicode.IsPunct) + ".."
						if float64(MeasureFont(sub, 12)) < (swidth - TextPadding) {
							break
						}
					}

					fmt.Fprintf(w, `<text style="font-size:12px;font-family:sans-serif;fill:#ffffff;" text-anchor="middle" x="%f" y="%d">%s</text>`, swidth/2.0, 4+DomainHeight/2, sub)
				}
			}
		}
		fmt.Fprintln(w, `</a></g>`)
	}

	if !*hideAxis {
		startY += DomainHeight + AxisPadding
		fmt.Fprintln(w, `<g class="axis">`)
		fmt.Fprintf(w, `<line x1="%d" x2="%d" y1="%d" y2="%d" stroke="#AAAAAA" />`, Padding, GraphicWidth-Padding, startY, startY)
		fmt.Fprintf(w, `<line x1="%d" x2="%d" y1="%d" y2="%d" stroke="#AAAAAA" />`, Padding, Padding, startY, startY+(AxisHeight/3))

		ts := TickSlice(ticks)
		sort.Sort(ts)
		lastDrawn := 0
		for i, t := range ts {
			if lastDrawn > 0 && (t.Pos-lastDrawn) < aaSpace {
				continue
			}
			j := ts.NextBetter(i, aaSpace)
			if i != j {
				continue
			}
			lastDrawn = t.Pos
			x := Padding + (float64(t.Pos) * scale)
			fmt.Fprintf(w, `<line x1="%f" x2="%f" y1="%d" y2="%d" stroke="#AAAAAA" />`, x, x, startY, startY+(AxisHeight/3))
			fmt.Fprintf(w, `<text style="font-size:10px;font-family:sans-serif;fill:#000000;" text-anchor="middle" x="%f" y="%d">%d</text>`, x, startY+AxisHeight, t.Pos)
		}

		fmt.Fprintln(w, "</g>")
	}

	fmt.Fprintln(w, svgFooter)
}
