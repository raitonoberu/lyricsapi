package spotify

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	ErrInvalidCookie = errors.New("invalid cookie provided")
	ErrNoToken       = errors.New("could not get token")
)

const (
	lyricsUrl = "https://spclient.wg.spotify.com/color-lyrics/v2/track"
	searchUrl = "https://api.spotify.com/v1/search"
	stateUrl  = "https://api.spotify.com/v1/me/player/currently-playing"
	userAgent = "Mozilla/5.0 (X11; Linux x86_64; rv:124.0) Gecko/20100101 Firefox/124.0"
)

const (
	lrclibUrl       = "https://lrclib.net/api/get?"
	lrclibUserAgent = "lyricsapi v0.1.0 (https://github.com/raitonoberu/lyricsapi)"
)

func NewClient(cookie string) *Client {
	return &Client{
		HttpClient: &http.Client{},
		cookie:     cookie,
	}
}

type Client struct {
	HttpClient *http.Client

	cookie string

	token    string
	tokenExp time.Time
	tokenMu  sync.Mutex
}

func (c *Client) GetByName(query string) ([]LyricsLine, error) {
	track, err := c.FindTrack(query)
	if err != nil {
		return nil, err
	}

	if track == nil {
		return nil, nil
	}

	if c.cookie != "" {
		// use spotify to get lyrics
		// if you don't have a premium account, it will only be first 5 lines
		return c.GetByID(track.ID)
	}
	// use lrclib to get lyrics by metadata
	// DEPRECATED and will be removed!
	return c.getFromLrclib(*track)
}

func (c *Client) GetByID(spotifyID string) ([]LyricsLine, error) {
	token, err := c.getToken()
	if err != nil {
		return nil, err
	}

	req, _ := http.NewRequest("GET", lyricsUrl+"/"+spotifyID, nil)
	req.Header = http.Header{
		"referer":             {"https://open.spotify.com/"},
		"origin":              {"https://open.spotify.com/"},
		"accept":              {"application/json"},
		"app-platform":        {"WebPlayer"},
		"spotify-app-version": {"1.2.61.20.g3b4cd5b2"},
		"user-agent":          {userAgent},
		"Authorization":       {"Bearer " + token},
	}
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	type responseType struct {
		Lyrics struct {
			Lines []LyricsLine `json:"lines"`
		} `json:"lyrics"`
	}

	var response responseType
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		if err == io.EOF {
			// this is thrown when the ID is invalid
			// or when the track has no lyrics
			return nil, nil
		}
		return nil, err
	}
	return response.Lyrics.Lines, nil
}

func (c *Client) FindTrack(query string) (*Track, error) {
	token, err := c.getToken()
	if err != nil {
		return nil, err
	}

	url := searchUrl + "?" + url.Values{
		"limit": {"1"},
		"type":  {"track"},
		"q":     {query},
	}.Encode()
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	type responseBody struct {
		Tracks struct {
			Items []*Track `json:"items"`
			Total int      `json:"total"`
		} `json:"tracks"`
	}

	var response responseBody
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	if response.Tracks.Total == 0 {
		return nil, nil
	}
	return response.Tracks.Items[0], nil
}

func (c *Client) State() (*StateResult, error) {
	token, err := c.getToken()
	if err != nil {
		return nil, err
	}

	req, _ := http.NewRequest("GET", stateUrl, nil)
	req.Header = http.Header{
		"Authorization": {"Bearer " + token},
	}
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response StateResult
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		if err == io.EOF {
			// stopped
			return nil, nil
		}
		return nil, err
	}
	return &response, nil
}

func (c *Client) getFromLrclib(t Track) ([]LyricsLine, error) {
	url := lrclibUrl + url.Values{
		"track_name":  {t.Name},
		"artist_name": {t.Artists[0].Name},
		"album_name":  {t.Album.Name},
		"duration":    {strconv.Itoa(t.Duration / 1000)},
	}.Encode()
	req, _ := http.NewRequest("GET", url, nil)
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response lrclibResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}
	return response.Parse(), nil
}

func (c *Client) getToken() (string, error) {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()

	if !c.tokenExpired() {
		return c.token, nil
	}

	if err := c.refreshToken(); err != nil {
		return "", fmt.Errorf("failed to refresh token: %w", err)
	}
	return c.token, nil
}

func (c *Client) tokenExpired() bool {
	return c.token == "" || time.Now().After(c.tokenExp)
}

type lrclibResponse struct {
	PlainLyrics  string `json:"plainLyrics"`
	SyncedLyrics string `json:"syncedLyrics"`
}

func (l *lrclibResponse) Parse() []LyricsLine {
	if l.SyncedLyrics != "" {
		return l.parceSynced()
	}
	if l.PlainLyrics != "" {
		return l.parsePlain()
	}
	return nil
}

func (l *lrclibResponse) parceSynced() []LyricsLine {
	lines := strings.Split(l.SyncedLyrics, "\n")
	result := make([]LyricsLine, len(lines))
	for i, line := range lines {
		result[i] = parseLrcLine(line)
	}
	return result
}

func (l *lrclibResponse) parsePlain() []LyricsLine {
	lines := strings.Split(l.PlainLyrics, "\n")
	result := make([]LyricsLine, len(lines))
	for i, line := range lines {
		result[i] = LyricsLine{Words: line}
	}
	return result
}

func parseLrcLine(line string) LyricsLine {
	// "[00:17.12] whatever"
	if len(line) < 11 {
		return LyricsLine{}
	}
	m, _ := strconv.Atoi(line[1:3])
	s, _ := strconv.Atoi(line[4:6])
	ms, _ := strconv.Atoi(line[7:9])
	return LyricsLine{
		Time:  m*60*1000 + s*1000 + ms*10,
		Words: line[11:],
	}
}
