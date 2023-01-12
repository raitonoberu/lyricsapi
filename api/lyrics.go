package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/raitonoberu/lyricsapi/lyrics"
)

var api = lyrics.NewLyricsApi(os.Getenv("COOKIE"))

// to change the json tag from "startTimeMs" to "time"
type lyricsLine struct {
	Time  int    `json:"time"`
	Words string `json:"words"`
}

func Lyrics(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	var lyrics *lyrics.LyricsResult
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

	if lrc, ok := query["lrc"]; ok && len(lrc) != 0 && lrc[0] == "1" {
		writeLrc(w, lyrics)
	} else {
		writeJson(w, lyrics)
	}

}

func writeJson(w http.ResponseWriter, lyrics *lyrics.LyricsResult) {
	w.Header().Set("content-type", "application/json; charset=utf-8")

	if lyrics != nil {
		w.Header().Set("Cache-Control", "s-maxage=86400")

		lines := make([]lyricsLine, len(lyrics.Lyrics.Lines))
		for i, l := range lyrics.Lyrics.Lines {
			lines[i] = lyricsLine(l)
		}
		json.NewEncoder(w).Encode(lines)
	} else {
		w.Write([]byte("[]"))
	}
}

func writeLrc(w http.ResponseWriter, lyrics *lyrics.LyricsResult) {
	w.Header().Set("content-type", "text/plain; charset=utf-8")

	if lyrics != nil {
		w.Header().Set("Cache-Control", "s-maxage=86400")

		lines := make([]string, len(lyrics.Lyrics.Lines))
		for i, l := range lyrics.Lyrics.Lines {
			lines[i] = fmt.Sprintf("[%d:%d.%d]%s", l.Time/60000, (l.Time/1000)%60, l.Time%1000, l.Words)
		}
		w.Write([]byte(strings.Join(lines, "\n")))
	} else {
		w.Write([]byte(""))
	}
}
