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
	"io/ioutil"

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
		err := LoadFontPath(path)
		if err == nil {
			return
		}
	}
}

func FontLoaded() bool {
	return fontContext != nil
}

func LoadFontPath(path string) error {
	fontBytes, err := ioutil.ReadFile(path)
	if err == nil {
		arialFont, err := freetype.ParseFont(fontBytes)
		if err == nil {
			fontContext = freetype.NewContext()
			fontContext.SetFont(arialFont)
		}
	}
	return err
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
