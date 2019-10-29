package main

import (
	"flag"
	"log"
	"net/http"
	"strings"
)

func main() {
	dir := flag.String("d", "./", "`directory` to serve")
	flag.Parse()
	fs := http.FileServer(http.Dir(*dir))

	log.Print("Open http://localhost:8080 in your web browser")
	http.ListenAndServe(":8080", http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Add("Cache-Control", "no-cache")
		if strings.HasSuffix(req.URL.Path, ".wasm") {
			resp.Header().Set("content-type", "application/wasm")
		}
		fs.ServeHTTP(resp, req)
	}))
}
