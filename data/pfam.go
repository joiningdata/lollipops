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
	"os"
)

const PfamGraphicURL = "http://pfam.xfam.org/protein/%s/graphic"

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
