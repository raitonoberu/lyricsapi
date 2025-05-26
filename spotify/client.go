package spotify

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/raitonoberu/lyricsapi/lyrics"
)

var (
	ErrInvalidCookie = errors.New("invalid cookie provided")
	ErrNoAccessToken = errors.New("could not get access token")
	ErrNoClientToken = errors.New("could not get client token")
)

const (
	lyricsUrl = "https://spclient.wg.spotify.com/color-lyrics/v2/track"
	searchUrl = "https://api.spotify.com/v1/search"
	stateUrl  = "https://api.spotify.com/v1/me/player/currently-playing"
	userAgent = "Mozilla/5.0 (X11; Linux x86_64; rv:124.0) Gecko/20100101 Firefox/124.0"
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

	clientId string

	accessToken    string
	accessTokenExp time.Time
	accessTokenMu  sync.Mutex

	clientToken    string
	clientTokenExp time.Time
	clientTokenMu  sync.Mutex
}

func (c *Client) GetByName(query string) ([]lyrics.Line, error) {
	track, err := c.FindTrack(query)
	if err != nil {
		return nil, err
	}

	if track == nil {
		return nil, nil
	}
	return c.GetByID(track.ID)
}

func (c *Client) GetByID(spotifyID string) ([]lyrics.Line, error) {
	accessToken, err := c.getAccessToken()
	if err != nil {
		return nil, err
	}
	clientToken, err := c.getClientToken()
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
		"Authorization":       {"Bearer " + accessToken},
		"client-token":        {clientToken},
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

	lines := make([]lyrics.Line, len(response.Lyrics.Lines))
	for i, l := range response.Lyrics.Lines {
		lines[i] = lyrics.Line(l)
	}
	return lines, nil
}

func (c *Client) FindTrack(query string) (*Track, error) {
	accessToken, err := c.getAccessToken()
	if err != nil {
		return nil, err
	}
	clientToken, err := c.getClientToken()
	if err != nil {
		return nil, err
	}

	url := searchUrl + "?" + url.Values{
		"limit": {"1"},
		"type":  {"track"},
		"q":     {query},
	}.Encode()
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("client-token", clientToken)
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
	accessToken, err := c.getAccessToken()
	if err != nil {
		return nil, err
	}
	clientToken, err := c.getClientToken()
	if err != nil {
		return nil, err
	}

	req, _ := http.NewRequest("GET", stateUrl, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("client-token", clientToken)
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

func (c *Client) getAccessToken() (string, error) {
	c.accessTokenMu.Lock()
	defer c.accessTokenMu.Unlock()

	if !c.accessTokenExpired() {
		return c.accessToken, nil
	}

	if err := c.refreshAccessToken(); err != nil {
		return "", fmt.Errorf("failed to refresh token: %w", err)
	}
	return c.accessToken, nil
}

func (c *Client) accessTokenExpired() bool {
	return c.accessToken == "" || time.Now().After(c.accessTokenExp)
}

func (c *Client) getClientToken() (string, error) {
	c.clientTokenMu.Lock()
	defer c.clientTokenMu.Unlock()

	if !c.clientTokenExpired() {
		return c.clientToken, nil
	}

	if err := c.refreshClientToken(); err != nil {
		return "", fmt.Errorf("failed to refresh token: %w", err)
	}
	return c.clientToken, nil
}

func (c *Client) clientTokenExpired() bool {
	return c.clientToken == "" || time.Now().After(c.clientTokenExp)
}
