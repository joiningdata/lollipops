package data

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
)

const UniprotDataURL = "https://www.uniprot.org/uniprot/%s.txt"

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

func GetUniprotGraphicData(accession string) (*GraphicResponse, error) {
	queryURL := fmt.Sprintf(UniprotDataURL, accession)
	resp, err := http.Get(queryURL)
	if err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			fmt.Fprintf(os.Stderr, "Unable to connect to Uniprot. Check your internet connection or try again later.")
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

	nouncertain := regexp.MustCompile("[?<>]")
	trimTags := regexp.MustCompile("[{][^}]*[}]")
	minisplit := regexp.MustCompile("[;.]")

	gd := &GraphicResponse{}
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
		case "FT":
			/// https://web.expasy.org/docs/userman.html#FT_line
			if strings.TrimSpace(line[:29]) == "" {
				// continuation of previous line's description (ignored)
				continue
			}
			featureType := strings.TrimSpace(line[:8])                                 //	Key name
			fromPos := strings.TrimSpace(nouncertain.ReplaceAllString(line[9:15], "")) //	'From' endpoint
			toPos := strings.TrimSpace(nouncertain.ReplaceAllString(line[16:22], ""))  //	'To' endpoint
			desc := strings.TrimSpace(line[29:])                                       //	Description

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
	}

	return gd, nil
}
