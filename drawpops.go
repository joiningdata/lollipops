package main

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"unicode"
)

const (
	LollipopRadius = 5
	LollipopHeight = 28
	BackboneHeight = 14
	DomainHeight   = 24
	Padding        = 15
	GraphicHeight  = LollipopRadius + LollipopHeight + BackboneHeight + DomainHeight + Padding*2
	//GraphicWidth   = 740
)

const svgHeader = `<?xml version='1.0'?>
<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="%d" height="%d">
<defs>
    <filter id="ds" x="0" y="0">
      <feOffset in="SourceAlpha" dx="2" dy="2" />
      <feComponentTransfer>
	    <feFuncA type="linear" slope="0.2"/>
	  </feComponentTransfer>
      <feGaussianBlur result="blurOut" stdDeviation="1" />
      <feBlend in="SourceGraphic" in2="blurOut" mode="normal" />
    </filter>
  </defs>
<style>text{font-size:12px;font-family:sans-serif;fill:#ffffff;}</style>
`
const svgFooter = `</svg>`

var stripChangePos = regexp.MustCompile("[A-Z][a-z]*([0-9]+)")

func DrawSVG(w io.Writer, GraphicWidth int, changelist []string, g *PfamGraphicResponse) {
	ht := GraphicHeight
	if len(changelist) == 0 {
		ht -= LollipopHeight
	}
	fmt.Fprintf(w, svgHeader, GraphicWidth, GraphicHeight)

	aaLen, _ := g.Length.Int64()
	scale := float64(GraphicWidth-Padding*2) / float64(aaLen)

	startY := Padding
	if len(changelist) > 0 {
		poptop := Padding + LollipopRadius
		popbot := poptop + LollipopHeight
		startY = popbot - (DomainHeight-BackboneHeight)/2

		// draw lollipops
		for _, chg := range changelist {
			cpos := stripChangePos.FindStringSubmatch(chg)
			spos := 0.0
			fmt.Sscanf(cpos[1], "%f", &spos)
			spos = Padding + (spos * scale)

			fmt.Fprintf(w, `<line x1="%f" x2="%f" y1="%d" y2="%d" stroke="#BABDB6" stroke-width="3"/>`, spos, spos, poptop, popbot)
			fmt.Fprintf(w, `<a xlink:title="%s"><circle cx="%f" cy="%d" r="%d" fill="#FF5555" /></a>`, chg, spos, poptop, LollipopRadius)
		}
	}

	// draw the backbone
	fmt.Fprintf(w, `<a xlink:title="%s, %s (%daa)"><rect fill="#BABDB6" x="%d" y="%d" width="%d" height="%d"/></a>`,
		g.Metadata.Identifier, g.Metadata.Description, aaLen,
		Padding, startY+(DomainHeight-BackboneHeight)/2, GraphicWidth-(Padding*2), BackboneHeight)

	// draw the curated domains
	for _, r := range g.Regions {
		sstart, _ := r.Start.Float64()
		swidth, _ := r.End.Float64()

		sstart *= scale
		swidth = (swidth * scale) - sstart

		fmt.Fprintf(w, `<g transform="translate(%f,%d)"><a xlink:href="%s" xlink:title="%s">`, Padding+sstart, startY, "http://pfam.xfam.org"+r.Link, r.Metadata.Description)
		fmt.Fprintf(w, `<rect fill="%s" x="0" y="0" width="%f" height="%d" filter="url(#ds)"/>`, r.Color, swidth, DomainHeight)
		if swidth > 40 {
			if len(r.Metadata.Description) > 1 && float64(len(r.Metadata.Description))*10 < swidth {
				// we can fit the full description! nice!
				fmt.Fprintf(w, `<text text-anchor="middle" x="%f" y="%d">%s</text>`, swidth/2.0, 4+DomainHeight/2, r.Metadata.Description)
			} else if float64(len(r.Text))*10 < swidth {
				fmt.Fprintf(w, `<text text-anchor="middle" x="%f" y="%d">%s</text>`, swidth/2.0, 4+DomainHeight/2, r.Text)
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
						if float64(len(pre+parts[i]+post))*10 < swidth {
							fmt.Fprintf(w, `<text text-anchor="middle" x="%f" y="%d">%s</text>`, swidth/2.0, 4+DomainHeight/2, pre+parts[i]+post)
							didOutput = true
							break
						}
						post = ".."
					}
				}

				if !didOutput {
					mx := int(swidth / 10)
					sub := strings.TrimFunc(r.Text[:mx], unicode.IsPunct) + ".."
					fmt.Fprintf(w, `<text text-anchor="middle" x="%f" y="%d">%s</text>`, swidth/2.0, 4+DomainHeight/2, sub)
				}
			}
		}
		fmt.Fprintln(w, `</a></g>`)
	}

	fmt.Fprintln(w, svgFooter)
}
