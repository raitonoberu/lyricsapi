package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/raitonoberu/lyricsapi/lyrics"
)

var api = lyrics.NewLyricsApi(os.Getenv("COOKIE"))

func Lyrics(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	var lyrics *lyrics.LyricsResult
	var err error
	if id := query.Get("id"); len(id) != 0 {
		log.Println("[INFO] Getting lyrics for ID:", id)
		lyrics, err = api.GetByID(id)
	} else if name := query.Get("name"); len(name) != 0 {
		log.Println("[INFO] Getting lyrics for query:", name)
		lyrics, err = api.GetByName(name)
	}

	if err != nil {
		log.Println("[ERROR]", err.Error())
	}

	writeHeader(w, query, lyrics, err)

	if query.Get("lrc") == "1" {
		writeLrc(w, lyrics)
	} else {
		writeJson(w, lyrics)
	}
}

func writeJson(w http.ResponseWriter, lyrics *lyrics.LyricsResult) {
	if lyrics == nil {
		w.Write([]byte("[]"))
		return
	}
	// [{"time":1000,"words":"words"}, ...]
	json.NewEncoder(w).Encode(lyrics.Lyrics.Lines)
}

func writeLrc(w http.ResponseWriter, lyrics *lyrics.LyricsResult) {
	if lyrics == nil {
		w.Write([]byte(""))
		return
	}

	lines := make([]string, len(lyrics.Lyrics.Lines))
	for i, l := range lyrics.Lyrics.Lines {
		lines[i] = fmt.Sprintf("[%02d:%02d.%02d]%s", l.Time/60000, (l.Time%60000)/1000, (l.Time%1000)/10, l.Words)
	}
	// [mm:ss.xx]words
	// ...
	w.Write([]byte(strings.Join(lines, "\n")))
}

func writeHeader(
	w http.ResponseWriter,
	query url.Values,
	lyrics *lyrics.LyricsResult,
	err error,
) {
	if query.Get("lrc") == "1" {
		w.Header().Set("content-type", "text/plain; charset=utf-8")
	} else {
		w.Header().Set("content-type", "application/json; charset=utf-8")
	}

	if err != nil {
		w.Header().Set("error", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Cache-Control", "s-maxage=86400")
	if lyrics == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Header().Set("color",
		fmt.Sprintf("#%X", lyrics.Colors.Background+0x1000000),
	)
	w.WriteHeader(http.StatusOK)
}
