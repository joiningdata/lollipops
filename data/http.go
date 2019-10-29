// +build !wasm

package data

import (
	"net/http"
	"net/url"
)

func httpGet(url string) (*http.Response, error) {
	return http.Get(url)
}

func httpPostForm(url string, vals url.Values) (*http.Response, error) {
	return http.PostForm(url, vals)
}
