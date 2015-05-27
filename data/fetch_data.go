package data

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// BiomartXML is the query used to translate HGNC Symbol into Uniprot/SwissProt Accession
const BiomartXML = `<!DOCTYPE Query><Query client="github.com/pbnjay/lollipops" processor="TSV" limit="-1" header="0">
	<Dataset name="hsapiens_gene_ensembl" config="gene_ensembl_config">
	<Filter name="with_uniprotswissprot" value="only" filter_list=""/>
	<Filter name="hgnc_symbol" value="%s" filter_list=""/>
	<Attribute name="uniprot_swissprot"/>
</Dataset></Query>`

const BiomartResultURL = "http://central.biomart.org/martservice/results"
const PfamGraphicURL = "http://pfam.xfam.org/protein/%s/graphic"

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

func GetProtID(symbol string) (string, error) {
	query := url.QueryEscape(fmt.Sprintf(BiomartXML, symbol))
	resp, err := http.Post(BiomartResultURL, "application/x-www-form-urlencoded",
		bytes.NewBufferString("query="+query))
	if err != nil {
		return "", err
	}
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("biomart error: %s", resp.Status)
	}
	respText := string(respBytes)
	return strings.TrimSpace(respText), nil
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
