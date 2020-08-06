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
	"os"
	"strings"
)

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

//GraphicMetaData this is a test ________________________________________
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
