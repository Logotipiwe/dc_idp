package main

import (
	"encoding/json"
	"errors"
	"fmt"
	env "github.com/logotipiwe/dc_go_env_lib"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type GoogleUser struct {
	Sub        string `json:"sub"`
	Name       string `json:"name"`
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
	Picture    string `json:"picture"`
	Locale     string `json:"locale"`
}

type DcUser struct {
	Id      string `json:"id"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

func getUserData(r *http.Request) (DcUser, error) {
	if os.Getenv("AUTO_AUTH") == "1" {
		return DcUser{
			Id:      os.Getenv("LOGOTIPIWE_GMAIL_ID"),
			Name:    "Reman Gerus",
			Picture: "https://cojo.ru/wp-content/uploads/2022/11/evaelfi-1-1.webp",
		}, nil
	}
	accessToken, err := getAccessTokenFromCookie(r)
	if err != nil {
		return DcUser{}, err
	}
	fmt.Println("Got AT from cookie, AT: " + accessToken)
	return getUserDataFromToken(accessToken)
}

func getAccessTokenFromCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie("access_token")
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func getUserDataFromToken(accessToken string) (DcUser, error) {
	bearer := "Bearer " + accessToken
	getUrl := "https://www.googleapis.com/oauth2/v3/userinfo"
	request, _ := http.NewRequest("GET", getUrl, nil)
	request.Header.Add("Authorization", bearer)

	client := &http.Client{}
	fmt.Println("Requesting user data...")
	res, err := client.Do(request)
	if err != nil {
		fmt.Println("Error requesting user data!")
		return DcUser{}, err
	}
	defer res.Body.Close()
	answerStr, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error reading user data!")
		return DcUser{}, err
	}
	fmt.Println("Answer with user data is:")
	fmt.Println(string(answerStr))
	var answer GoogleUser
	err = json.Unmarshal(answerStr, &answer)
	fmt.Printf("User data is %s %s %s", answer.Sub, answer.Name, answer.Picture)
	if answer.Sub == "" {
		return DcUser{}, errors.New(string(answerStr))
	}
	return DcUser{
		Id:      answer.Sub,
		Name:    answer.Name,
		Picture: answer.Picture,
	}, nil
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
