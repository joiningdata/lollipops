// +build !wasm

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/pbnjay/lollipops/data"
	"github.com/pbnjay/lollipops/drawing"
)

func createOutput(filename string, d *data.GraphicResponse, variants []string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stderr, "Drawing diagram to", filename)
	if strings.HasSuffix(strings.ToLower(filename), ".png") {
		drawing.DrawPNG(f, *dpi, variants, d)
	} else {
		drawing.DrawSVG(f, variants, d)
	}
	return f.Close()
}
