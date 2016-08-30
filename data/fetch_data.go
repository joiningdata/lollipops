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
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const PfamGraphicURL = "http://pfam.xfam.org/protein/%s/graphic"

// PfamMotifNames has human-readable names from http://pfam.xfam.org/help#tabview=tab9
var PfamMotifNames = map[string]string{
	"disorder":       "Disordered region (Pfam/IUPred)",
	"low_complexity": "Low complexity region (Pfam/SEG)",
	"sig_p":          "Signal peptide region (Pfam/Phobius)",
	"coiled_coil":    "Coiled-coil motif (Pfam/ncoils)",
	"transmembrane":  "Transmembrane region (Pfam/Phobius)",
}

// PfamGraphicFeature is a generic representation of various Pfam feature responses
type PfamGraphicFeature struct {
	Color         string              `json:"colour"`
	StartStyle    string              `json:"startStyle"`
	EndStyle      string              `json:"endStyle"`
	Text          string              `json:"text"`
	Type          string              `json:"type"`
	Start         json.Number         `json:"start"`
	End           json.Number         `json:"end"`
	ShouldDisplay bool                `json:"display"`
	Link          string              `json:"href"`
	Metadata      PfamGraphicMetadata `json:"metadata"`
	// many unused fields...
}

type PfamGraphicMetadata struct {
	Accession   string `json:"accession"`
	Description string `json:"description"`
	Identifier  string `json:"identifier"`
}

type PfamGraphicResponse struct {
	Length   json.Number          `json:"length"`
	Markups  []PfamGraphicFeature `json:"markups"`
	Metadata PfamGraphicMetadata  `json:"metadata"`
	Motifs   []PfamGraphicFeature `json:"motifs"`
	Regions  []PfamGraphicFeature `json:"regions"`
}

func GetPfamGraphicData(accession string) (*PfamGraphicResponse, error) {
	queryURL := fmt.Sprintf(PfamGraphicURL, accession)
	resp, err := http.Get(queryURL)
	if err != nil {
		return nil, err
	}
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("pfam error: %s", resp.Status)
	}

	data := []PfamGraphicResponse{}
	err = json.Unmarshal(respBytes, &data)
	//if err != nil {
	//	return nil, err
	//}
	if len(data) != 1 {
		return nil, fmt.Errorf("pfam returned invalid result")
	}
	r := data[0]
	return &r, nil
}

func GetProtID(symbol string) (string, error) {
	apiURL := `http://www.uniprot.org/uniprot/?query=` + url.QueryEscape(symbol)
	apiURL += `+AND+reviewed:yes+AND+organism:9606+AND+database:pfam`
	apiURL += `&sort=score&columns=id,entry+name,reviewed,genes,organism&format=tab`

	resp, err := http.Get(apiURL)
	if err != nil {
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
	if nmatches > 1 {
		fmt.Fprintf(os.Stderr, "Uniprot returned %d hits for your gene symbol '%s':\n", nmatches, symbol)
		fmt.Fprintln(os.Stderr, string(respBytes))
	}
	if bestHit == 0 {
		log.Fatalf("Unable to find protein ID for '%s'", symbol)
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
