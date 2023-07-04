package main

import (
	"encoding/json"
	"errors"
	"fmt"
	env "github.com/logotipiwe/dc_go_env_lib"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type User struct {
	Sub        string `json:"sub"`
	Name       string `json:"name"`
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
	Picture    string `json:"picture"`
	Locale     string `json:"locale"`
}

func getUserData(r *http.Request) (User, error) {
	accessToken, err := getAccessTokenFromCookie(r)
	if err != nil {
		return User{}, err
	}
	return getUserDataFromToken(accessToken)
}

func getAccessTokenFromCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie("access_token")
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func getUserDataFromToken(accessToken string) (User, error) {
	bearer := "Bearer " + accessToken
	getUrl := "https://www.googleapis.com/oauth2/v3/userinfo"
	request, _ := http.NewRequest("GET", getUrl, nil)
	request.Header.Add("Authorization", bearer)

	client := &http.Client{}
	res, _ := client.Do(request)
	defer res.Body.Close()
	var answer User
	err := json.NewDecoder(res.Body).Decode(&answer)
	if err != nil {
		return User{}, err
	}
	if answer.Sub != "" {
		return answer, nil
	} else {
		return answer, errors.New("WTF HUH")
	}
}

func exchangeCodeToToken(code string) string {
	println(fmt.Sprintf("Exchanging code %v to token", code))
	postUrl := "https://oauth2.googleapis.com/token"
	data := url.Values{}
	data.Set("client_id", clientId)
	data.Set("client_secret", os.Getenv("G_OAUTH_CLIENT_SECRET"))
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	wtfIsThat := env.GetCurrUrl() + env.GetSubpath() + "/g-oauth"
	data.Set("redirect_uri", wtfIsThat)
	client := &http.Client{}
	req, _ := http.NewRequest(http.MethodPost, postUrl, strings.NewReader(data.Encode())) // URL-encoded payload
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	println(fmt.Sprintf("Getting AT from url %v", postUrl))
	println(fmt.Sprintf("Redirect url (for what?) %v", wtfIsThat))
	resp, _ := client.Do(req)
	defer resp.Body.Close()
	var answer map[string]string
	json.NewDecoder(resp.Body).Decode(&answer)
	if resp.StatusCode != 200 {
		println(fmt.Sprintf("Got error while exchanging code to token. Status: %d. Body: %v", resp.StatusCode, answer))
	}
	accessToken := answer["access_token"]
	println(fmt.Sprintf("Access token got with status 200. Token = %v", accessToken))
	return accessToken
}

func getRedirectUri(r *http.Request) string {
	redirToCookie, err := r.Cookie(redirAfterAuthCookieName)
	var redirTo string
	if err != nil {
		println("Error getting redirectTo cookie")
		redirTo = ""
	}
	redirTo = redirToCookie.Value

	if redirTo == "" {
		redirTo = getFallbackRedirect()
		println(fmt.Sprintf(
			"RedirTo value is empty. Setting fallback redirect %v", redirTo))
	}
	return redirTo
}
