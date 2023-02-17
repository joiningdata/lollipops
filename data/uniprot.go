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
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

const UniprotDataURL = "https://rest.uniprot.org/uniprotkb/%s.txt"

var defaultUniprotFeatures = map[string][]string{
	"COILED":   {"motif", "coiled_coil", "#9cff00"},
	"SIGNAL":   {"motif", "sig_p", "#ff9c00"},
	"TRANSMEM": {"motif", "transmembrane", "#ff0000"},
	"COMPBIAS": {"motif", "low_complexity", "#00ffff"},

	"DNA_BIND": {"region", "dna_bind", "#ff5353"},
	"ZN_FING":  {"region", "zn_fing", "#2dcf00"},
	"CA_BIND":  {"region", "ca_bind", "#86bcff"},

	"MOTIF":  {"region", "motif", "#1fc01f"},
	"REPEAT": {"region", "repeat", "#1fc01f"},
	"DOMAIN": {"region", "domain", "#9999ff"},
}

func getValueForKey(line, key string) string {
	parts := strings.Split(line, ";")
	for _, s := range parts {
		p := strings.SplitN(s, "=", 2)
		if p[0] == key {
			return strings.TrimSpace(p[1])
		}
	}
	return ""
}

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

func GetUniprotGraphicData(accession string) (*GraphicResponse, error) {
	queryURL := fmt.Sprintf(UniprotDataURL, accession)
	resp, err := httpGet(queryURL)
	if err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			fmt.Fprintf(os.Stderr, "Unable to connect to Uniprot. Check your internet connection or try again later.")
			os.Exit(1)
		}
		return nil, err
	}
	defer resp.Body.Close()
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	respBytes = uniprotDecompress(respBytes)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("pfam error: %s", resp.Status)
	}
	trimTags := regexp.MustCompile("[{][^}]*[}]")
	minisplit := regexp.MustCompile("[;.]")

	gd := &GraphicResponse{}
	pat := regexp.MustCompile(`FT\s+([A-Z_]+)\s+(\d+)\.\.(\d+)\nFT\s+\/note="([\w\s\d]+)"`)
	matches := pat.FindAllSubmatch(respBytes, -1)

	for _, match := range matches {
		featureType := string(match[1])
		fromPos := string(match[2]) //	'From' endpoint
		toPos := string(match[3])   //	'To' endpoint
		desc := string(match[4])    //	Description
		if fromPos == "" || toPos == "" || fromPos == toPos {
			// skip any unknown positions or point features
			continue
		}
		fdata, ok := defaultUniprotFeatures[featureType]
		if !ok {
			continue
		}
		desc = strings.TrimSpace(trimTags.ReplaceAllString(desc, ""))
		shortDesc := desc
		if p := minisplit.Split(desc, 2); len(p) == 2 {
			shortDesc = strings.TrimSpace(p[0])
		}

		feat := GraphicFeature{
			Color: fdata[2],
			Text:  strings.Trim(shortDesc, ". "),
			Type:  fdata[1],
			Start: json.Number(fromPos),
			End:   json.Number(toPos),
			Metadata: GraphicMetadata{
				Description: strings.Trim(shortDesc, ". "),
			},
		}
		switch fdata[0] {
		case "region":
			gd.Regions = append(gd.Regions, feat)
		case "motif":
			gd.Motifs = append(gd.Motifs, feat)
		default:
			log.Println("unknown feature set", fdata[0])
		}
	}

	for _, bline := range bytes.Split(respBytes, []byte("\n")) {
		if len(bline) < 5 {
			continue
		}
		key := string(bytes.TrimSpace(bline[:5]))
		line := string(bline[5:])
		switch key {
		case "GN":
			// GN   Name=CTNNB1; Synonyms=CTNNB; ORFNames=OK/SW-cl.35, PRO2286;
			sym := getValueForKey(line, "Name")
			if sym != "" {
				gd.Metadata.Identifier = sym
			}
		case "DE":
			// DE   RecName: Full=Catenin beta-1;
			if !strings.HasPrefix(line, "RecName: ") {
				continue
			}
			desc := getValueForKey(line[9:], "Full")
			if desc != "" {
				gd.Metadata.Description = desc
			}
		case "SQ":
			// SQ   SEQUENCE   781 AA;  85497 MW;  CB78F165A3EEF86E CRC64;
			parts := strings.Split(line, ";")
			for _, p := range parts {
				if !strings.HasPrefix(p, "SEQUENCE") {
					continue
				}
				seqLen := strings.TrimSpace(strings.TrimSuffix(p[8:], "AA"))
				gd.Length = json.Number(seqLen)
				break
			}
			////////////////////////////
		}
	}

	return gd, nil
}

const UNIPROTRESTURL = "https://rest.uniprot.org/uniprotkb/search?query=%s+AND+reviewed:true+AND+organism_id:9606&columns=id,entry+name,reviewed,genes,organism&format=tsv"

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
	for _, line := range strings.Split(string(respBytes), "\n") {
		p := strings.Split(string(line), "\t")
		for _, g := range strings.Split(string(p[4]), " ") {
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
