# Lollipops web frontend

This directory contains a simple web frontend for lollipops, using the
template provided by Go's wasm toolchain in `$GOROOT/misc/wasm/`

Building lollipops for wasm using:

	cd $GOPATH/github.com/joiningdata/lollipops
	GOOS=js GOARCH=wasm go build -o html/lollipops.wasm .

You can then copy the 4 files to any webserver:

	index.html
	lollipops.wasm
	style.css
	wasm_exec.js


A simple server is included here for testing purposes:

	cd $GOPATH/github.com/joiningdata/lollipops/html
	go run www.go
