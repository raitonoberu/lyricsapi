package main

import (
	"net/http"

	"github.com/raitonoberu/lyricsapi/api"
)

func main() {
	http.HandleFunc("/api/lyrics", api.Lyrics)
	http.ListenAndServe(":8080", nil)
}
