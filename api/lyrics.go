package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/raitonoberu/lyricsapi/lyrics"
)

var api = lyrics.NewLyricsApi(os.Getenv("COOKIE"))

type Result struct {
	Lines []*lyricsLine `json:"lines,omitempty"`
	Error string        `json:"error,omitempty"`
}

type lyricsLine struct {
	Time  int64  `json:"time"`
	Words string `json:"words"`
}

func Lyrics(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	var lyrics *lyrics.ColorLyrics
	var err error
	if id, ok := query["id"]; ok && len(id) != 0 {
		lyrics, err = api.Get(id[0])
	} else if name, ok := query["name"]; ok && len(name) != 0 {
		lyrics, err = api.GetByName(name[0])
	}

	result := &Result{}
	statusCode := 404

	if err != nil {
		result.Error = err.Error()
		log.Println(err.Error())
		statusCode = 500
	}
	if lyrics != nil {
		result.Lines = make([]*lyricsLine, len(lyrics.Lyrics.Lines))
		for i, l := range lyrics.Lyrics.Lines {
			result.Lines[i] = &lyricsLine{
				Time:  l.StartTimeMs,
				Words: l.Words,
			}
		}
		statusCode = 200
	}

	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(result)
}
