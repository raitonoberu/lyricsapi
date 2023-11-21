package lyrics

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"time"
)

var (
	ErrInvalidCookie = errors.New("invalid cookie provided")
	ErrNoToken       = errors.New("could not get token")
)

const tokenUrl = "https://open.spotify.com/get_access_token?reason=transport&productType=web_player"
const lyricsUrl = "https://spclient.wg.spotify.com/color-lyrics/v2/track/"
const searchUrl = "https://api.spotify.com/v1/search?"

func NewLyricsApi(cookie string) *LyricsApi {
	return &LyricsApi{
		Client: http.DefaultClient,
		cookie: cookie,
	}
}

type LyricsApi struct {
	Client *http.Client

	cookie    string
	token     string
	expiresIn time.Time
}

func (l *LyricsApi) GetByName(query string) (*LyricsResult, error) {
	if err := l.checkToken(); err != nil {
		return nil, err
	}

	url := searchUrl + url.Values{
		"limit": {"1"},
		"type":  {"track"},
		"q":     {query},
	}.Encode()
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+l.token)
	resp, err := l.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result := &searchResult{}
	err = json.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		return nil, err
	}
	if result.Tracks.Total == 0 {
		return nil, nil
	}

	return l.GetByID(result.Tracks.Items[0].ID)
}

func (l *LyricsApi) GetByID(spotifyID string) (*LyricsResult, error) {
	if err := l.checkToken(); err != nil {
		return nil, err
	}

	req, _ := http.NewRequest("GET", lyricsUrl+spotifyID, nil)
	req.Header = http.Header{
		"referer":          {"https://open.spotify.com/"},
		"origin":           {"https://open.spotify.com/"},
		"accept":           {"application/json"},
		"accept-language":  {"en"},
		"app-platform":     {"WebPlayer"},
		"sec-ch-ua-mobile": {"?0"},

		"Authorization": {"Bearer " + l.token},
	}
	resp, err := l.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result := &LyricsResult{}
	err = json.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		if err == io.EOF {
			// this is thrown when the ID is invalid
			// or when the track has no lyrics
			return nil, nil
		}
		return nil, err
	}
	return result, nil
}

func (l *LyricsApi) checkToken() error {
	if !l.tokenExpired() {
		return nil
	}
	return l.updateToken()
}

func (l *LyricsApi) tokenExpired() bool {
	return l.token == "" || time.Now().After(l.expiresIn)
}

func (l *LyricsApi) updateToken() error {
	req, _ := http.NewRequest("GET", tokenUrl, nil)
	req.Header = http.Header{
		"referer":             {"https://open.spotify.com/"},
		"origin":              {"https://open.spotify.com/"},
		"accept":              {"application/json"},
		"accept-language":     {"en"},
		"app-platform":        {"WebPlayer"},
		"sec-fetch-dest":      {"empty"},
		"sec-fetch-mode":      {"cors"},
		"sec-fetch-site":      {"same-origin"},
		"spotify-app-version": {"1.1.54.35.ge9dace1d"},
		"cookie":              {l.cookie},
	}
	resp, err := l.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	result := &tokenBody{}
	err = json.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		return err
	}

	if result.IsAnonymous {
		return ErrInvalidCookie
	}
	if result.AccessToken == "" {
		return ErrNoToken
	}

	l.token = result.AccessToken
	l.expiresIn = time.Unix(0, result.ExpiresIn*int64(time.Millisecond))
	return nil
}

type searchResult struct {
	Tracks struct {
		Items []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"items"`
		Total int `json:"total"`
	} `json:"tracks"`
}

type tokenBody struct {
	AccessToken string `json:"accessToken"`
	ExpiresIn   int64  `json:"accessTokenExpirationTimestampMs"`
	IsAnonymous bool   `json:"isAnonymous"`
}
