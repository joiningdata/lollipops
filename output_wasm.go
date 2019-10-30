// +build wasm

package main

import (
	"bytes"
	"fmt"
	"os"
	"syscall/js"

	"github.com/joiningdata/lollipops/data"
	"github.com/joiningdata/lollipops/drawing"
)

func createOutput(elementID string, d *data.GraphicResponse, variants []string) error {
	fmt.Fprintln(os.Stderr, "Creating SVG image")
	buf := &bytes.Buffer{}
	drawing.DrawSVG(buf, variants, d)
	js.Global().Get("document").Call("getElementById", "lollipops-svg-container").Set("innerHTML", buf.String())

	return nil
}
