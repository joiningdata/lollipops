package drawing

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"code.google.com/p/jamslam-freetype-go/freetype"
)

var (
	arialPath   *string
	fontContext *freetype.Context
)

func init() {
	// try to find Arial so we can measure it
	// I don't try very hard...
	popularpaths := []string{
		// OS X path
		"/Library/Fonts/Arial.ttf",

		// Windows path
		"C:/Windows/Fonts/arial.ttf",

		// Ubuntu with multiverse msttcorefonts package
		"/usr/share/fonts/truetype/msttcorefonts/arial.ttf",
	}
	for _, path := range popularpaths {
		fontBytes, err := ioutil.ReadFile(path)
		if err == nil {
			arialFont, err := freetype.ParseFont(fontBytes)
			if err == nil {
				fontContext = freetype.NewContext()
				fontContext.SetFont(arialFont)
				return
			}
		}
	}

	fmt.Fprintln(os.Stderr, "can't find arial.ttf - for more accurate font sizing use -f=/path/to/arial.ttf")
	arialPath = flag.String("f", "", "path to arial.ttf")
}

// MeasureFont returns the pixel width of the string s at font size sz.
// It tries to use system Arial font if possible, but falls back to a
// conservative ballpark estimate otherwise.
func MeasureFont(s string, sz int) int {
	// use actual TTF font metrics if available
	if fontContext != nil {
		fontContext.SetFontSize(float64(sz))
		w, _, _ := fontContext.MeasureString(s)
		return freetype.Pixel(w)
	}

	return len(s) * (sz - 2)
}
