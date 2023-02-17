//go:build !wasm
// +build !wasm

package data

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

func httpGet(url string) (*http.Response, error) {
	return http.Get(url)
}

func httpGetInsecure(url string) (*http.Response, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	fmt.Fprintln(os.Stderr, "WARNING: making insecure request to ", url)
	fmt.Fprintln(os.Stderr, "         eventually this will no longer work correctly!")
	return client.Get(url)
}

func httpPostForm(url string, vals url.Values) (*http.Response, error) {
	return http.PostForm(url, vals)
}
