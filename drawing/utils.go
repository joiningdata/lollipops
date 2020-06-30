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
	"math"
	"regexp"
	"strings"

	"github.com/joiningdata/lollipops/data"
)

var stripChangePos = regexp.MustCompile("(^|[A-Za-z]*)([0-9]+)([A-Za-z]*)")

type Tick struct {
	Pos int
	Pri int
	Cnt int
	Col string

	isLollipop bool
	label      string
	x          float64
	y          float64
	r          float64
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

// BlendColorStrings blends two CSS #RRGGBB colors together with a straight average.
func BlendColorStrings(a, b string) string {
	var r1, g1, b1, r2, g2, b2 int
	fmt.Sscanf(strings.ToUpper(a), "#%02X%02X%02X", &r1, &g1, &b1)
	fmt.Sscanf(strings.ToUpper(b), "#%02X%02X%02X", &r2, &g2, &b2)
	return fmt.Sprintf("#%02X%02X%02X", (r1+r2)/2, (g1+g2)/2, (b1+b2)/2)
}

// AutoWidth automatically determines the best width to use to fit all
// available domain names into the plot.
func (s *Settings) AutoWidth(g *data.GraphicResponse) float64 {
	aaLen, _ := g.Length.Float64()
	w := 400.0
	if s.dpi != 0 {
		w *= s.dpi / 72.0
	}

	for _, r := range g.Regions {
		sstart, _ := r.Start.Float64()
		send, _ := r.End.Float64()

		aaPart := (send - sstart) / aaLen
		minTextWidth := float64(s.MeasureFont(r.Text, 12)) + (s.TextPadding * 2) + 1

		ww := minTextWidth / aaPart
		if ww > w {
			w = ww
		}
	}
	return w + (s.Padding * 2)
}

func (t *Tick) Radius(s *Settings) float64 {
	if t.Cnt <= 1 {
		return s.LollipopRadius
	}
	return math.Sqrt(math.Log(float64(2+t.Cnt)) * s.LollipopRadius * s.LollipopRadius)
}
