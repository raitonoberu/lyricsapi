package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/raitonoberu/lyricsapi/itunes"
	"github.com/raitonoberu/lyricsapi/lrclib"
	"github.com/raitonoberu/lyricsapi/lyrics"
	"github.com/raitonoberu/lyricsapi/spotify"
)

var spotifyApi = spotify.NewClient(os.Getenv("COOKIE"))

func Lyrics(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	var lyrics []lyrics.Line
	var err error
	if id := query.Get("id"); len(id) != 0 {
		log.Println("[INFO] Getting lyrics for ID:", id)
		lyrics, err = spotifyApi.GetByID(id)
	} else if name := query.Get("name"); len(name) != 0 {
		log.Println("[INFO] Getting lyrics for query:", name)
		var track *itunes.Track
		track, err = itunes.Search(name)
		if track != nil {
			lyrics, err = lrclib.GetLyrics(lrclib.GetLyricsRequest{
				TrackName:  track.TrackName,
				ArtistName: track.ArtistName,
				AlbumName:  track.CollectionName,
			})
		}
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

func writeJson(w io.Writer, lyrics []lyrics.Line) {
	if lyrics == nil {
		w.Write([]byte("[]"))
		return
	}
	// [{"time":1000,"words":"words"}, ...]
	json.NewEncoder(w).Encode(lyrics)
}

func writeLrc(w io.Writer, lyrics []lyrics.Line) {
	if lyrics == nil {
		w.Write([]byte(""))
		return
	}

	lines := make([]string, len(lyrics))
	for i, l := range lyrics {
		// [mm:ss.xx]words
		lines[i] = fmt.Sprintf("[%02d:%02d.%02d]%s", l.Time/60000, (l.Time%60000)/1000, (l.Time%1000)/10, l.Words)
	}
	w.Write([]byte(strings.Join(lines, "\n")))
}

func writeHeader(
	w http.ResponseWriter,
	query url.Values,
	lyrics []lyrics.Line,
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
	if len(lyrics) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
}
