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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const PfamGraphicURL = "http://pfam.xfam.org/protein/%s/graphic"

// MotifNames has human-readable names
//  - mostly from http://pfam.xfam.org/help#tabview=tab9
var MotifNames = map[string]string{
	"disorder":       "Disordered region",
	"low_complexity": "Low complexity region",
	"sig_p":          "Signal peptide region",
	"coiled_coil":    "Coiled-coil motif",
	"transmembrane":  "Transmembrane region",
}

// GraphicFeature is a generic representation of various feature responses
type GraphicFeature struct {
	Color    string          `json:"colour"`
	Text     string          `json:"text"`
	Type     string          `json:"type"`
	Start    json.Number     `json:"start"`
	End      json.Number     `json:"end"`
	Link     string          `json:"href"`
	Metadata GraphicMetadata `json:"metadata"`
}

type GraphicMetadata struct {
	Description string `json:"description"`
	Identifier  string `json:"identifier"`
}

type GraphicResponse struct {
	Length   json.Number      `json:"length"`
	Metadata GraphicMetadata  `json:"metadata"`
	Motifs   []GraphicFeature `json:"motifs"`
	Regions  []GraphicFeature `json:"regions"`
}

func GetLocalGraphicData(filename string) (*GraphicResponse, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	pf := &GraphicResponse{}
	err = json.NewDecoder(f).Decode(pf)
	f.Close()
	for i, x := range pf.Motifs {
		if x.Link != "" && !strings.Contains(x.Link, "://") {
			x.Link = "http://pfam.xfam.org" + x.Link
			pf.Motifs[i] = x
		}
	}
	for i, x := range pf.Regions {
		if x.Link != "" && !strings.Contains(x.Link, "://") {
			x.Link = "http://pfam.xfam.org" + x.Link
			pf.Regions[i] = x
		}
	}
	return pf, err
}

func GetPfamGraphicData(accession string) (*GraphicResponse, error) {
	queryURL := fmt.Sprintf(PfamGraphicURL, accession)
	resp, err := http.Get(queryURL)
	if err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			fmt.Fprintf(os.Stderr, "Unable to connect to Pfam. Check your internet connection or try again later.")
			os.Exit(1)
		}
		return nil, err
	}
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("pfam error: %s", resp.Status)
	}

	data := []GraphicResponse{}
	err = json.Unmarshal(respBytes, &data)
	//if err != nil {
	//	return nil, err
	//}
	if len(data) != 1 {
		return nil, fmt.Errorf("pfam returned invalid result")
	}
	r := data[0]
	for i, x := range r.Motifs {
		if x.Link != "" {
			x.Link = "http://pfam.xfam.org" + x.Link
			r.Motifs[i] = x
		}
	}
	for i, x := range r.Regions {
		if x.Link != "" {
			x.Link = "http://pfam.xfam.org" + x.Link
			r.Regions[i] = x
		}
	}
	return &r, nil
}

func GetProtID(symbol string) (string, error) {
	apiURL := `http://www.uniprot.org/uniprot/?query=` + url.QueryEscape(symbol)
	apiURL += `+AND+reviewed:yes+AND+organism:9606+AND+database:pfam`
	apiURL += `&sort=score&columns=id,entry+name,reviewed,genes,organism&format=tab`

	resp, err := http.Get(apiURL)
	if err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			fmt.Fprintf(os.Stderr, "Unable to connect to Uniprot. Check your internet connection or try again later.")
			os.Exit(1)
		}
		return "", err
	}
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("uniprot error: %s", resp.Status)
	}
	nmatches := 0
	bestHit := 0
	protID := ""
	for _, line := range strings.Split(string(respBytes), "\n") {
		n := strings.Count(line, symbol)
		if n >= bestHit {
			p := strings.SplitN(line, "\t", 4)
			if len(p) < 4 {
				continue
			}
			for _, g := range strings.Split(p[3], " ") {
				if g == symbol {
					// exact match, return immediately
					return p[0], nil
				}
			}
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

func GetProtMapping(dbname, geneid string) (string, error) {
	apiURL := `http://www.uniprot.org/mapping/`
	params := url.Values{
		"from":   {dbname},
		"query":  {geneid}, // wish i could filter only reviewed:yes here...
		"to":     {"ACC"},
		"format": {"tab"},
	}

	resp, err := http.PostForm(apiURL, params)
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
