//go:build wasm
// +build wasm

package data

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// implements http.Get but makes wasm's fetch work with CORS
func httpGet(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("js.fetch:mode", "cors")

	resp, err := http.DefaultClient.Do(req)
	bb, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	resp.Body = ioutil.NopCloser(strings.NewReader(string(bb)))
	return resp, err
}

func httpGetInsecure(url string) (*http.Response, error) {
	// I am not willing to test if this can be configured in WASM.
	return httpGet(url)
}

// implements http.PostForm but makes wasm's fetch work with CORS
func httpPostForm(wurl string, vals url.Values) (*http.Response, error) {
	body := strings.NewReader(vals.Encode())
	req, err := http.NewRequest("POST", wurl, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("js.fetch:mode", "cors")

	resp, err := http.DefaultClient.Do(req)
	bb, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	resp.Body = ioutil.NopCloser(strings.NewReader(string(bb)))
	return resp, err
}
