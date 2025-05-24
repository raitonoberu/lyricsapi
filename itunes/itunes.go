package itunes

import (
	"encoding/json"
	"net/http"
	"net/url"
)

func Search(query string) (*Track, error) {
	res, err := http.Get("https://itunes.apple.com/search?" + url.Values{
		"term":  {query},
		"media": {"music"},
		"limit": {"5"},
	}.Encode())
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	ret := SearchResult{}
	err = json.NewDecoder(res.Body).Decode(&ret)
	if err != nil {
		return nil, err
	}

	for _, t := range ret.Results {
		if t.Kind == "song" {
			return &t, nil
		}
	}

	return nil, nil
}

type Track struct {
	TrackName      string `json:"trackName"`
	ArtistName     string `json:"artistName"`
	CollectionName string `json:"collectionName"`
	Kind           string `json:"kind"`
}

type SearchResult struct {
	Results []Track `json:"results"`
}
