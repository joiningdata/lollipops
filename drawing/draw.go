package drawing

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/pbnjay/lollipops/data"
)

type diagram struct {
	*Settings

	g          *data.PfamGraphicResponse
	changelist []string

	ticks        TickSlice
	domainLabels []string
	startY       float64
}

func (s *Settings) prepare(changelist []string, g *data.PfamGraphicResponse) *diagram {
	d := &diagram{
		Settings:   s,
		g:          g,
		changelist: changelist,
	}
	if s.GraphicWidth == 0 {
		s.GraphicWidth = s.AutoWidth(g)
	}
	aaLen, _ := g.Length.Int64()
	scale := (s.GraphicWidth - s.Padding*2) / float64(aaLen)
	popSpace := int((s.LollipopRadius + 2) / scale)
	startY := s.Padding
	if s.ShowLabels {
		startY += s.Padding // add some room for labels
	}

	pops := TickSlice{}
	col := s.SynonymousColor
	s.GraphicHeight = s.DomainHeight + s.Padding*2
	if len(changelist) > 0 {
		popMatch := make(map[string]int)
		// parse changelist and check if lollipops need staggered
		for i, chg := range changelist {
			if chg == "" {
				continue
			}
			cnt := 1
			cpos := stripChangePos.FindStringSubmatch(chg)
			spos := 0
			col = s.SynonymousColor
			if len(cpos) == 4 && (cpos[3] != "" && cpos[3] != "=" && cpos[3] != cpos[1]) {
				col = s.MutationColor
			}
			if strings.Contains(chg, "@") {
				parts := strings.SplitN(chg, "@", 2)
				fmt.Sscanf(parts[1], "%d", &cnt)
				chg = parts[0]
			}
			if strings.Contains(chg, "#") {
				parts := strings.SplitN(chg, "#", 2)
				col = "#" + parts[1]
				chg = parts[0]
			}
			changelist[i] = chg
			fmt.Sscanf(cpos[2], "%d", &spos)
			col = strings.ToLower(col)
			if idx, f := popMatch[chg+col]; f {
				pops[idx].Cnt += cnt
			} else {
				popMatch[chg+col] = len(pops)
				pops = append(pops, Tick{Pos: spos, Pri: -i, Cnt: cnt, Col: col})
			}
		}
		sort.Sort(pops)
		maxStaggered := s.LollipopRadius + s.LollipopHeight
		for pi, pop := range pops {
			h := s.LollipopRadius + s.LollipopHeight
			for pj := pi + 1; pj < len(pops); pj++ {
				if pops[pj].Pos-pop.Pos > popSpace {
					break
				}
				h += 0.5 + (pop.Radius(s) * 3.0)
			}
			if h > maxStaggered {
				maxStaggered = h
			}
		}
		s.GraphicHeight += maxStaggered
		startY += maxStaggered - (s.LollipopRadius + s.LollipopHeight)
	}
	if !s.HideAxis {
		s.GraphicHeight += s.AxisPadding + s.AxisHeight
	}

	d.startY = startY
	d.ticks = append(d.ticks,
		Tick{Pos: 0, Pri: 0},           // start isn't very important (0 is implied)
		Tick{Pos: int(aaLen), Pri: 99}, // always draw the length in the axis
	)

	if len(pops) > 0 {
		poptop := startY + s.LollipopRadius
		popbot := poptop + s.LollipopHeight
		startY = popbot - (s.DomainHeight-s.BackboneHeight)/2

		// position lollipops
		for pi, pop := range pops {
			spos := s.Padding + (float64(pop.Pos) * scale)
			mytop := poptop
			for pj := pi + 1; pj < len(pops); pj++ {
				if pops[pj].Pos-pop.Pos > popSpace {
					break
				}
				mytop -= 0.5 + (pops[pj].Radius(s) * 3.0)
			}

			d.ticks = append(d.ticks, Tick{
				Pos: pop.Pos,
				Pri: 10,
				Col: pop.Col,

				isLollipop: true,
				label:      changelist[-pop.Pri],
				x:          spos,
				y:          mytop,
				r:          pop.Radius(s),
			})
		}
	}

	if !s.HideMotifs {
		// if motifs are shown, add ticks as necessary
		for _, r := range g.Motifs {
			if r.Type == "pfamb" {
				continue
			}
			if r.Type != "disorder" {
				tstart, _ := r.Start.Int64()
				tend, _ := r.End.Int64()
				d.ticks = append(d.ticks, Tick{Pos: int(tstart), Pri: 1})
				d.ticks = append(d.ticks, Tick{Pos: int(tend), Pri: 1})
			}
		}
	}

	// determine labels for the curated domains
	for _, r := range g.Regions {
		sstart, _ := r.Start.Float64()
		swidth, _ := r.End.Float64()

		d.ticks = append(d.ticks, Tick{Pos: int(sstart), Pri: 5})
		d.ticks = append(d.ticks, Tick{Pos: int(swidth), Pri: 5})

		sstart *= scale
		swidth = (swidth * scale) - sstart

		label := ""

		if swidth > 10 {
			if len(r.Metadata.Description) > 1 && float64(MeasureFont(r.Metadata.Description, 12)) < (swidth-s.TextPadding) {
				// we can fit the full description! nice!
				label = r.Metadata.Description
			} else if float64(MeasureFont(r.Text, 12)) < (swidth - s.TextPadding) {
				label = r.Text
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
						if float64(MeasureFont(pre+parts[i]+post, 12)) < (swidth - s.TextPadding) {
							label = pre + parts[i] + post
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
						if float64(MeasureFont(sub, 12)) < (swidth - s.TextPadding) {
							break
						}
					}

					label = sub
				}
			}
		}

		d.domainLabels = append(d.domainLabels, label)
	}

	sort.Sort(d.ticks)
	return d
}
