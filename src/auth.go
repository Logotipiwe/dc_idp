package main

import (
	"encoding/json"
	"fmt"
	"net/http"
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
	Id       string `json:"id"`
	Name     string `json:"name"`
	Picture  string `json:"picture"`
	GoogleId string `json:"googleId"`
}

func getUserData(r *http.Request) (*DcUser, error) {
	accessToken, err := getAccessTokenFromCookie(r)
	if err != nil {
		fmt.Println(err) //only log because test user without AT can be returned later
	}
	fmt.Println("Got google AT from cookie, AT: " + accessToken)
	gUser, err := getGoogleUserDataFromGoogleAT(accessToken)
	if err != nil {
		return nil, err
	}
	user, err := getUserFromDbByGoogleId(gUser.Sub)
	if err == nil {
		userJson, _ := json.Marshal(user)
		fmt.Println(string(userJson))
	}
	return user, err
}

func getAccessTokenFromCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie("access_token")
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
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
