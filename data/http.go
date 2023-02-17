// +build !wasm

package data

import (
	"net/http"
	"net/url"
    "crypto/tls"
)

func httpGet(url string) (*http.Response, error) {
	tr := &http.Transport{
        TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
    }
    client := &http.Client{Transport: tr}
    return client.Get(url)
}

func httpPostForm(url string, vals url.Values) (*http.Response, error) {
	return http.PostForm(url, vals)
}
