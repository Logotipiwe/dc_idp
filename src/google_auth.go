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

func getGoogleUserDataFromGoogleAT(googleAccessToken string) (*DcUser, error) {
	if isGoogleAutoAuth() {
		autoUser := getAutoAuthedUser()
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
	return &DcUser{
		Id:       answer.Sub,
		Name:     answer.Name,
		Picture:  answer.Picture,
		GoogleId: answer.Sub, //TODO temp
	}, nil
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

func createGoogleUserIfNeeded(user *DcUser) error {
	exists, err := existsInDbByGoogleId(user.Id)
	if err != nil {
		fmt.Println(err)
		return err
	}
	if !exists {
		fmt.Println("User with google id " + user.GoogleId + " doesn't exist. Creating in db...")
		user, err = createUserInDb(user)
		if err != nil {
			return err
		}
		fmt.Println("User with google id " + user.GoogleId + " created!")
	}
	return nil
}
