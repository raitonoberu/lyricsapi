package spotify

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/pquerna/otp/totp"
)

const (
	serverTimeUrl = "https://open.spotify.com/server-time"
	tokenUrl      = "https://open.spotify.com/get_access_token"
)

// let's hope this lasts
const secret = "GU2TANZRGQ2TQNJTGQ4DONBZHE2TSMRSGQ4DMMZQGMZDSMZUG4"

func (c *Client) refreshToken() error {
	totp, totpTime, err := c.getTotp()
	if err != nil {
		return fmt.Errorf("failed to get totp: %w", err)
	}
	timeStr := fmt.Sprint(totpTime.Unix())

	url := tokenUrl + "?" + url.Values{
		"reason":      {"transport"},
		"productType": {"web-player"},
		"totp":        {totp},
		"totpServer":  {totp},
		"totpVer":     {"5"},
		"sTime":       {timeStr},
		"cTime":       {timeStr + "420"},
	}.Encode()

	req, _ := http.NewRequest("GET", url, nil)
	req.Header = http.Header{
		"referer":             {"https://open.spotify.com/"},
		"origin":              {"https://open.spotify.com/"},
		"accept":              {"application/json"},
		"app-platform":        {"WebPlayer"},
		"spotify-app-version": {"1.2.61.20.g3b4cd5b2"},
		"user-agent":          {userAgent},
		"cookie":              {c.cookie},
	}
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	type responseType struct {
		AccessToken string `json:"accessToken"`
		ExpiresIn   int64  `json:"accessTokenExpirationTimestampMs"`
		IsAnonymous bool   `json:"isAnonymous"`
	}

	var response responseType
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return err
	}

	if c.cookie != "" && response.IsAnonymous {
		return ErrInvalidCookie
	}
	if response.AccessToken == "" {
		return ErrNoToken
	}

	c.token = response.AccessToken
	c.tokenExp = time.Unix(0, response.ExpiresIn*int64(time.Millisecond))
	return nil
}

func (c *Client) getTotp() (string, time.Time, error) {
	serverTime, err := c.getServerTime()
	if err != nil {
		// fuck it we ball
		serverTime = time.Now()
	}
	totp, err := totp.GenerateCode(secret, serverTime)
	if err != nil {
		return "", time.Time{}, err
	}
	return totp, serverTime, nil
}

func (c *Client) getServerTime() (time.Time, error) {
	req, _ := http.NewRequest("GET", serverTimeUrl, nil)
	req.Header = http.Header{
		"referer":             {"https://open.spotify.com/"},
		"origin":              {"https://open.spotify.com/"},
		"accept":              {"application/json"},
		"app-platform":        {"WebPlayer"},
		"spotify-app-version": {"1.2.61.20.g3b4cd5b2"},
		"user-agent":          {userAgent},
		"cookie":              {c.cookie},
	}
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return time.Time{}, err
	}
	defer resp.Body.Close()

	type responseType struct {
		ServerTime int64 `json:"serverTime"`
	}

	var response responseType
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return time.Time{}, err
	}
	return time.Unix(response.ServerTime, 0), nil
}
