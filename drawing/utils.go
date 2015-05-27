package drawing

import (
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/pbnjay/lollipops/data"
)

var stripChangePos = regexp.MustCompile("(^|[A-Za-z]*)([0-9]+)([A-Za-z]*)")

type Tick struct {
	Pos int
	Pri int
	Cnt int
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

// BlendColorStrings blends two CSS #RRGGBB colors together with a straight average.
func BlendColorStrings(a, b string) string {
	var r1, g1, b1, r2, g2, b2 int
	fmt.Sscanf(strings.ToUpper(a), "#%02X%02X%02X", &r1, &g1, &b1)
	fmt.Sscanf(strings.ToUpper(b), "#%02X%02X%02X", &r2, &g2, &b2)
	return fmt.Sprintf("#%02X%02X%02X", (r1+r2)/2, (g1+g2)/2, (b1+b2)/2)
}

// AutoWidth automatically determines the best width to use to fit all
// available domain names into the plot.
func AutoWidth(g *data.PfamGraphicResponse) int {
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

func (t *Tick) Radius() float64 {
	if t.Cnt <= 1 {
		return LollipopRadius
	}
	return math.Sqrt(math.Log(float64(2+t.Cnt)) * LollipopRadius * LollipopRadius)
}
