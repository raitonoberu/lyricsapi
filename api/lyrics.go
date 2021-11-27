package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/raitonoberu/lyricsapi/lyrics"
)

var api = lyrics.NewLyricsApi(os.Getenv("COOKIE"))

type lyricsLine struct {
	Time  int64  `json:"time"`
	Words string `json:"words"`
}

func Lyrics(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	var lyrics *lyrics.ColorLyrics
	var err error
	if id, ok := query["id"]; ok && len(id) != 0 {
		log.Println("[INFO] Getting lyrics for ID", id)
		lyrics, err = api.Get(id[0])
		if err != nil {
			log.Println("[ERROR]", err.Error(), id)
		}
	} else if name, ok := query["name"]; ok && len(name) != 0 {
		log.Println("[INFO] Getting lyrics for query", name)
		lyrics, err = api.GetByName(name[0])
		if err != nil {
			log.Println("[ERROR]", err.Error(), name)
		}
	}

	if lyrics == nil {
		fmt.Fprint(w, "[]")
		return
	}

	lines := make([]*lyricsLine, len(lyrics.Lyrics.Lines))
	for i, l := range lyrics.Lyrics.Lines {
		lines[i] = &lyricsLine{
			Time:  l.StartTimeMs,
			Words: l.Words,
		}
	}
	json.NewEncoder(w).Encode(lines)
}
