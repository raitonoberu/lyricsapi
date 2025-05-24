package lrclib

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/raitonoberu/lyricsapi/lyrics"
)

const (
	lrclibUrl       = "https://lrclib.net/api/get?"
	lrclibUserAgent = "lyricsapi v0.1.0 (https://github.com/raitonoberu/lyricsapi)"
)

type GetLyricsRequest struct {
	TrackName  string
	ArtistName string
	AlbumName  string
}

func GetLyrics(r GetLyricsRequest) ([]lyrics.Line, error) {
	url := lrclibUrl + url.Values{
		"track_name":  {r.TrackName},
		"artist_name": {r.ArtistName},
		"album_name":  {r.AlbumName},
		// "duration":    {strconv.Itoa(r.Duration / 1000)},
	}.Encode()
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response response
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}
	return parseResponse(response), nil
}

type response struct {
	PlainLyrics  string `json:"plainLyrics"`
	SyncedLyrics string `json:"syncedLyrics"`
}

func parseResponse(r response) []lyrics.Line {
	if r.SyncedLyrics != "" {
		return parceSynced(r)
	}
	if r.PlainLyrics != "" {
		return parsePlain(r)
	}
	return nil
}

func parceSynced(r response) []lyrics.Line {
	lines := strings.Split(r.SyncedLyrics, "\n")
	result := make([]lyrics.Line, len(lines))
	for i, line := range lines {
		result[i] = parseLrcLine(line)
	}
	return result
}

func parsePlain(r response) []lyrics.Line {
	lines := strings.Split(r.PlainLyrics, "\n")
	result := make([]lyrics.Line, len(lines))
	for i, line := range lines {
		result[i] = lyrics.Line{Words: line}
	}
	return result
}

func parseLrcLine(line string) lyrics.Line {
	// "[00:17.12] whatever"
	if len(line) < 11 {
		return lyrics.Line{}
	}
	m, _ := strconv.Atoi(line[1:3])
	s, _ := strconv.Atoi(line[4:6])
	ms, _ := strconv.Atoi(line[7:9])
	return lyrics.Line{
		Time:  m*60*1000 + s*1000 + ms*10,
		Words: line[11:],
	}
}
