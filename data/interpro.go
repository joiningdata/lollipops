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
	"os"
	"sort"
)

const InterProURL = "https://www.ebi.ac.uk/interpro/api/entry/%s/protein/uniprot/%s/?extra_fields=short_name&page_size=100"
const InterProLink = "https://www.ebi.ac.uk/interpro/entry/%s/%s"
const SequenceFeaturesURL = "https://www.ebi.ac.uk/interpro/api/protein/UniProt/%s/?extra_features=true"

func GetProteinMatches(database string, accession string) ([]GraphicFeature, error) {
	var sourceDatabase string
	filterDomains := false
	if database == "interpro" {
		sourceDatabase = "all"
		filterDomains = true
	} else {
		sourceDatabase = "pfam"
	}
	queryURL := fmt.Sprintf(InterProURL, sourceDatabase, accession)
	resp, err := httpGet(queryURL)
	if err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			fmt.Fprintf(os.Stderr, "Unable to connect to InterPro. Check your internet connection or try again later.")
			os.Exit(1)
		}
		return nil, err
	}
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("InterPro error: %s", resp.Status)
	}

	r := InterProEntryResponse{}
	err = json.Unmarshal(respBytes, &r)
	if err != nil {
		return nil, err
	}

	var gs []GraphicFeature
	for _, e := range r.Entries {
		for _, m := range e.Matches {
			for _, l := range m.Locations {
				for _, f := range l.Fragments {
					if !filterDomains || f.Representative {
						gf := GraphicFeature{
							Text:  e.ExtraFields.ShortName,
							Type:  e.Metadata.Type,
							Start: f.Start,
							End:   f.End,
							Link:  fmt.Sprintf(InterProLink, e.Metadata.Database, e.Metadata.Accession),
							Metadata: GraphicMetadata{
								Description: e.Metadata.Name,
								Identifier:  e.Metadata.Accession,
							},
						}
						gs = append(gs, gf)
					}
				}
			}
		}
	}

	sort.Slice(gs, func(i, j int) bool {
		start1, _ := gs[i].Start.Int64()
		start2, _ := gs[j].Start.Int64()

		if start1 != start2 {
			return start1 < start2
		}

		end1, _ := gs[i].End.Int64()
		end2, _ := gs[j].End.Int64()
		return end1 < end2
	})

	hexColors := [14]string{
		"#2DCF00", "#FF5353", "#5B5BFF", "#EBD61D", "#BA21E0", "#FF9C42", "#FF7DFF",
		"#B9264F", "#BABA21", "#C48484", "#1F88A7", "#CAFEB8", "#4A9586", "#CEB86C",
	}

	for i := 0; i < len(gs); i++ {
		gs[i].Color = hexColors[i%len(hexColors)]
	}

	return gs, nil
}

func GetSequenceFeatures(accession string) ([]GraphicFeature, error) {
	queryURL := fmt.Sprintf(SequenceFeaturesURL, accession)
	resp, err := httpGet(queryURL)
	if err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			fmt.Fprintf(os.Stderr, "Unable to connect to InterPro. Check your internet connection or try again later.")
			os.Exit(1)
		}
		return nil, err
	}
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("InterPro error: %s", resp.Status)
	}

	data := make(map[string]InterProFeature)

	err = json.Unmarshal(respBytes, &data)
	if err != nil {
		return nil, fmt.Errorf("InterPro error: %s", err)
	}

	var gs []GraphicFeature
	featureDatabases := map[string]string{
		"signalp_e":  "sig_p",
		"signalp_g+": "sig_p",
		"signalp_g-": "sig_p",
		"coils":      "coiled_coil",
		"tmhmm":      "transmembrane",
	}
	for _, feature := range data {
		if feature.Database == "mobidblt" {
			for _, location := range feature.Locations {
				for _, fragment := range location.Fragments {
					if fragment.SeqFeature == "Consensus Disorder Prediction" {
						gf := GraphicFeature{
							Color: "#CCCCCC",
							Type:  "disorder",
							Start: fragment.Start,
							End:   fragment.End,
						}
						gs = append(gs, gf)
					}
				}
			}

			continue
		}

		for feature_db, feature_type := range featureDatabases {
			if feature.Database == feature_db {

				for _, location := range feature.Locations {
					for _, fragment := range location.Fragments {
						gf := GraphicFeature{
							Color: "#CCCCCC",
							Type:  feature_type,
							Start: fragment.Start,
							End:   fragment.End,
						}
						gs = append(gs, gf)
					}
				}

				break
			}
		}
	}

	return gs, nil
}
