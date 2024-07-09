package travellermap

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

func Search(query string) (*SearchResults, error) {
	resp, err := http.Get(fmt.Sprintf("https://travellermap.com/api/search?q=%s", query))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("%d response from travellermap", resp.StatusCode))
	}
	body, _ := io.ReadAll(io.Reader(resp.Body))
	var results SearchResults
	if err = json.Unmarshal(body, &results); err != nil {
		return nil, err
	}
	return &results, nil
}

type SearchResults struct {
	Results struct {
		Count int `json:"Count"`
		Items []struct {
			World *struct {
				HexX       int    `json:"HexX"`
				HexY       int    `json:"HexY"`
				Sector     string `json:"Sector"`
				Uwp        string `json:"Uwp"`
				SectorX    int    `json:"SectorX"`
				SectorY    int    `json:"SectorY"`
				Name       string `json:"Name"`
				SectorTags string `json:"SectorTags"`
			} `json:"World,omitempty"`
			Label *struct {
				HexX       int    `json:"HexX"`
				HexY       int    `json:"HexY"`
				Scale      int    `json:"Scale"`
				SectorX    int    `json:"SectorX"`
				SectorY    int    `json:"SectorY"`
				Name       string `json:"Name"`
				SectorTags string `json:"SectorTags"`
			} `json:"Label,omitempty"`
		} `json:"Items"`
	} `json:"Results"`
}

func FetchWorldDetail(sector string, hex string) (*WorldDetail, error) {
	resp, err := http.Get(fmt.Sprintf("https://travellermap.com/data/%s/%s", sector, hex))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("%d response from travellermap", resp.StatusCode))
	}
	body, _ := io.ReadAll(io.Reader(resp.Body))
	var results WorldResults
	if err = json.Unmarshal(body, &results); err != nil {
		return nil, err
	}

	if len(results.Worlds) > 0 {
		return &results.Worlds[0], nil
	}
	return nil, errors.New(fmt.Sprintf("No results on travellermap for %s/%s", sector, hex))
}

type WorldResults struct {
	Worlds []WorldDetail `json:"Worlds"`
}

type WorldDetail struct {
	Name               string `json:"Name"`
	Hex                string `json:"Hex"`
	Uwp                string `json:"UWP"`
	Pbg                string `json:"PBG"`
	Zone               string `json:"Zone"`
	Bases              string `json:"Bases"`
	Allegiance         string `json:"Allegiance"`
	Stellar            string `json:"Stellar"`
	Ss                 string `json:"SS"`
	Ix                 string `json:"Ix"`
	Ex                 string `json:"Ex"`
	Cx                 string `json:"Cx"`
	Nobility           string `json:"Nobility"`
	Worlds             int    `json:"Worlds"`
	ResourceUnits      int    `json:"ResourceUnits"`
	Subsector          int    `json:"Subsector"`
	Quadrant           int    `json:"Quadrant"`
	WorldX             int    `json:"WorldX"`
	WorldY             int    `json:"WorldY"`
	Remarks            string `json:"Remarks"`
	LegacyBaseCode     string `json:"LegacyBaseCode"`
	Sector             string `json:"Sector"`
	SubsectorName      string `json:"SubsectorName"`
	SectorAbbreviation string `json:"SectorAbbreviation"`
	AllegianceName     string `json:"AllegianceName"`
}
