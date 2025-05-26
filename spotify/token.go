package spotify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/pquerna/otp/totp"
)

const (
	serverTimeUrl  = "https://open.spotify.com/server-time"
	accessTokenUrl = "https://open.spotify.com/get_access_token"
	clientTokenUrl = "https://clienttoken.spotify.com/v1/clienttoken"
)

// let's hope this lasts
const secret = "GU2TANZRGQ2TQNJTGQ4DONBZHE2TSMRSGQ4DMMZQGMZDSMZUG4"

func (c *Client) refreshAccessToken() error {
	totp, totpTime, err := c.getTotp()
	if err != nil {
		return fmt.Errorf("failed to get totp: %w", err)
	}
	timeStr := fmt.Sprint(totpTime.Unix())

	url := accessTokenUrl + "?" + url.Values{
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
		ClientId    string `json:"clientId"`
		AccessToken string `json:"accessToken"`
		ExpiresIn   int64  `json:"accessTokenExpirationTimestampMs"`
		IsAnonymous bool   `json:"isAnonymous"`
	}

	var response responseType
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return err
	}

	if response.IsAnonymous {
		return ErrInvalidCookie
	}
	if response.AccessToken == "" {
		return ErrNoAccessToken
	}

	c.clientId = response.ClientId
	c.accessToken = response.AccessToken
	c.accessTokenExp = time.Unix(0, response.ExpiresIn*int64(time.Millisecond))
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

func (c *Client) refreshClientToken() error {
	_, err := c.getAccessToken()
	if err != nil {
		return err
	}

	type clientData struct {
		ClientVersion string   `json:"client_version"`
		ClientId      string   `json:"client_id"`
		JsSdkData     struct{} `json:"js_sdk_data"` // leaving this empty seems to work
	}

	type requestBody struct {
		ClientData clientData `json:"client_data"`
	}

	buf := bytes.Buffer{}
	if err := json.NewEncoder(&buf).Encode(requestBody{
		ClientData: clientData{
			ClientVersion: "1.2.65.120.g1371365b",
			ClientId:      c.clientId,
		},
	}); err != nil {
		return err
	}

	req, _ := http.NewRequest("POST", clientTokenUrl, &buf)
	req.Header = http.Header{
		"referer":    {"https://open.spotify.com/"},
		"origin":     {"https://open.spotify.com/"},
		"accept":     {"application/json"},
		"user-agent": {userAgent},
	}
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	type responseBody struct {
		Type         string `json:"response_type"`
		GrantedToken struct {
			Token        string `json:"token"`
			ExpiresAfter int    `json:"expires_after_seconds"`
			RefreshAfter int    `json:"refresh_after_seconds"`
		} `json:"granted_token"`
	}

	var response responseBody
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return err
	}

	if response.Type != "RESPONSE_GRANTED_TOKEN_RESPONSE" || response.GrantedToken.Token == "" {
		return ErrNoClientToken
	}

	expiresIn := response.GrantedToken.RefreshAfter
	if expiresIn == 0 {
		expiresIn = response.GrantedToken.ExpiresAfter
	}

	c.clientToken = response.GrantedToken.Token
	c.clientTokenExp = time.Now().Add(time.Duration(expiresIn) * time.Second)
	return nil
}
