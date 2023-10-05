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

package data

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func uniprotDecompress(respBytes []byte) []byte {
	// uniprot's REST implementation doesn't set a valid Content-Encoding header when
	// gzipping the response, so Go's automatic gzip decompression doesn't work.
	// since they'll probably fix it after put in this workaround, we'll just try
	// to un-gzip and replace the content if it doesn't fail.

	buf := bytes.NewReader(respBytes)
	zrdr, err := gzip.NewReader(buf)
	if err != nil {
		return respBytes
	}
	data, err := io.ReadAll(zrdr)
	if err == nil {
		return data
	}
	return respBytes
}

const UNIPROTRESTURL = "https://rest.uniprot.org/uniprotkb/search?query=%s+AND+reviewed:true+AND+organism_id:9606&format=tsv&fields=accession,gene_names,length"

func GetProtID(symbol string) (string, error) {
	apiURL := fmt.Sprintf(UNIPROTRESTURL, symbol)
	resp, err := http.Get(apiURL)
	if err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			fmt.Fprintf(os.Stderr, "Unable to connect to Uniprot. Check your internet connection or try again later.")
			os.Exit(1)
		}
		return "", err
	}
	defer resp.Body.Close()
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	respBytes = uniprotDecompress(respBytes)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("uniprot error: %s", resp.Status)
	}
	nmatches := 0
	bestHit := 0
	protID := ""
	for i, line := range strings.Split(string(respBytes), "\n") {
		if i == 0 {
			continue
		}
		p := strings.Split(string(line), "\t")
		for _, g := range strings.Split(string(p[1]), " ") {
			if g == symbol {
				// exact match, return immediately
				return p[0], nil
			}
		}
		n := strings.Count(line, symbol)
		if n >= bestHit {
			bestHit = n
			protID = p[0]
		}
		nmatches++
	}
	fmt.Fprintf(os.Stderr, "Uniprot returned %d hits for your gene symbol '%s':\n", nmatches, symbol)
	if nmatches > 1 {
		fmt.Fprintln(os.Stderr, string(respBytes))
	}
	if bestHit == 0 {
		fmt.Fprintf(os.Stderr, "Unable to find protein ID for '%s' (use -U XX to select one of the above)\n", symbol)
		os.Exit(1)
	} else if nmatches > 1 {
		fmt.Fprintf(os.Stderr, "Selected '%s' as the best match. Use -U XXX to use another ID.\n\n", protID)
	}
	return protID, nil
}

func GetProtLength(accession string) (int, error) {
	apiURL := fmt.Sprintf("https://rest.uniprot.org/uniprotkb/%s.json", accession)
	resp, err := http.Get(apiURL)
	if err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			fmt.Fprintf(os.Stderr, "Unable to connect to Uniprot. Check your internet connection or try again later.")
			os.Exit(1)
		}
		return 0, err
	}
	defer resp.Body.Close()
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	respBytes = uniprotDecompress(respBytes)
	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("uniprot error: %s", resp.Status)
	}

	data := UniProtResponse{}
	err = json.Unmarshal(respBytes, &data)
	if err != nil {
		return 0, err
	}

	return data.Sequence.Length, nil
}

func GetProtMapping(dbname, geneid string) (string, error) {
	apiURL := `https://www.uniprot.org/uploadlists/`
	params := url.Values{
		"from":   {dbname},
		"query":  {geneid}, // wish i could filter only reviewed:yes here...
		"to":     {"ACC"},
		"format": {"tab"},
	}

	resp, err := httpPostForm(apiURL, params)
	if err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			fmt.Fprintf(os.Stderr, "Unable to connect to Uniprot. Check your internet connection or try again later.")
			os.Exit(1)
		}
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("uniprot error: %s", resp.Status)
	}
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var res []string
	protID := ""
	for i, line := range strings.Split(string(respBytes), "\n") {
		if i == 0 { //skip header
			continue
		}
		p := strings.SplitN(line, "\t", 2)
		if len(p) == 2 {
			res = append(res, p[1])
			// take the shortest acc in the hopes it's reviewed
			if protID == "" || len(p[1]) < len(protID) {
				protID = p[1]
			}
		}
	}
	if len(res) > 1 {
		fmt.Println("More than one Uniprot result: ", strings.Join(res, ", "))
	}
	return protID, nil
}
