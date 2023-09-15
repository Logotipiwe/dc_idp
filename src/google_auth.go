package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	env "github.com/logotipiwe/dc_go_env_lib"
	"github.com/logotipiwe/dc_go_utils/src/config"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func getGoogleUserDataFromGoogleAT(googleAccessToken string) (*GoogleUser, error) {
	if isGoogleAutoAuth() {
		fmt.Println("Auto auth enabled, getting auto authed user...")
		autoUser := getAutoAuthedGoogleUser()
		userJson, _ := json.Marshal(autoUser)
		fmt.Println(string(userJson))
		return &autoUser, nil
	}
	bearer := "Bearer " + googleAccessToken
	getUrl := "https://www.googleapis.com/oauth2/v3/userinfo"
	request, _ := http.NewRequest("GET", getUrl, nil)
	request.Header.Add("Authorization", bearer)

	client := &http.Client{}
	fmt.Println("Requesting user data...")
	res, err := client.Do(request)
	if err != nil {
		fmt.Println("Error requesting user data!")
		return nil, err
	}
	defer res.Body.Close()
	answerStr, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error reading user data!")
		return nil, err
	}
	fmt.Println("Answer with user data is:")
	fmt.Println(string(answerStr))
	var answer GoogleUser
	err = json.Unmarshal(answerStr, &answer)
	fmt.Printf("User data is %s %s %s", answer.Sub, answer.Name, answer.Picture)
	if answer.Sub == "" {
		return nil, errors.New(string(answerStr))
	}
	return &answer, nil
}

func getAutoAuthedGoogleUser() GoogleUser {
	return GoogleUser{
		Sub:        config.GetConfig("LOGOTIPIWE_GMAIL_ID"),
		Name:       "Reman Gerus",
		GivenName:  "Reman",
		FamilyName: "Gerus",
		Picture:    "https://cojo.ru/wp-content/uploads/2022/11/evaelfi-1-1.webp",
		Locale:     "??",
	}
}

func exchangeGoogleCodeToToken(code string) string {
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

func createDcUserFromGoogleUser(gUser *GoogleUser) *DcUser {
	var dcId string
	if gUser.Sub == config.GetConfig("LOGOTIPIWE_GMAIL_ID") {
		dcId = config.GetConfig("LOGOTIPIWE_DC_ID")
	} else {
		dcId = uuid.NewString()
	}
	return &DcUser{
		Id:       dcId,
		Name:     gUser.Name,
		Picture:  gUser.Picture,
		GoogleId: gUser.Sub,
	}
}
