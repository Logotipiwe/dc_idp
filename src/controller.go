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

const clientId = "319710408255-ntkf14k8ruk4p98sn2u1ho4j99rpjqja.apps.googleusercontent.com"
const redirAfterAuthCookieName = "redirect_after_auth"

func main() {
	http.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		redirect := r.URL.Query().Get("redirect")
		println("/auth")
		println(redirect)
		redirectUrl := env.GetCurrUrl() + env.GetSubpath() + "/g-oauth"
		googleOauthUri := "https://accounts.google.com/o/oauth2/v2/auth?" +
			"client_id=" + url.QueryEscape(clientId) +
			"&redirect_uri=" + url.QueryEscape(redirectUrl) +
			"&response_type=code&scope=profile"
		println(googleOauthUri)
		setRedirectAfterAuthCookie(w, redirect)
		http.Redirect(w, r, googleOauthUri, 302)
	})

	http.HandleFunc("/g-oauth", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		println("Code is: " + code)
		token := exchangeCodeToToken(r, code)
		setATCookie(w, token)
		println(token)

		redirToCookie, err := r.Cookie(redirAfterAuthCookieName)
		var redirTo string
		if err != nil {
			redirTo = ""
		}
		redirTo = redirToCookie.Value
		if redirTo == "" {
			redirTo = "/"
		}
		http.Redirect(w, r, redirTo, 302)
	})

	http.HandleFunc("/default", func(w http.ResponseWriter, r *http.Request) {
		//setATCookie(w, "")
		//toIndex(w, r)
		println("/default")
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")
		data, err := getUserData(r)
		if err != nil {
			fmt.Fprint(w, "те бы авторизоваться <a href='"+env.GetSubpath()+"/auth'>вот тут типа</a>")
		} else {
			fmt.Fprint(w, "Здарова "+data.Name+"!")
		}
	})

	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		setATCookie(w, "")
		toIndex(w, r)
	})

	err := http.ListenAndServe(":"+os.Getenv("CONTAINER_PORT"), nil)
	println("Server up!")
	if err != nil {
		panic("Lol server fell")
	}
}

func exchangeCodeToToken(r *http.Request, code string) string {
	postUrl := "https://oauth2.googleapis.com/token"
	data := url.Values{}
	data.Set("client_id", clientId)
	data.Set("client_secret", os.Getenv("G_OAUTH_CLIENT_SECRET"))
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", env.GetCurrUrl()+env.GetSubpath()+"/g_oauth")
	client := &http.Client{}
	req, _ := http.NewRequest(http.MethodPost, postUrl, strings.NewReader(data.Encode())) // URL-encoded payload
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, _ := client.Do(req)
	defer resp.Body.Close()
	var answer map[string]string
	json.NewDecoder(resp.Body).Decode(&answer)
	if resp.StatusCode != 200 {
		fmt.Printf("Got error while exchanging code to token. Status: %d. Body: %v", resp.StatusCode, answer)
	}
	return answer["access_token"]
}

func setRedirectAfterAuthCookie(w http.ResponseWriter, destination string) {
	cookie := http.Cookie{
		Name:     redirAfterAuthCookieName,
		Value:    destination,
		HttpOnly: false,
	}
	http.SetCookie(w, &cookie)
}

func setATCookie(w http.ResponseWriter, token string) {
	cookie := http.Cookie{
		Name:     "access_token",
		Value:    token,
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)
}

func toIndex(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/default", 302)
}

func getUserData(r *http.Request) (*User, error) {
	cookie, _ := r.Cookie("access_token")
	var accessToken string
	if cookie != nil {
		accessToken = cookie.Value
	} else {
		return nil, errors.New("empty access token")
	}
	bearer := "Bearer " + accessToken
	getUrl := "https://www.googleapis.com/oauth2/v3/userinfo"
	request, _ := http.NewRequest("GET", getUrl, nil)
	request.Header.Add("Authorization", bearer)

	client := &http.Client{}
	res, _ := client.Do(request)
	defer res.Body.Close()
	var answer User
	json.NewDecoder(res.Body).Decode(&answer)
	if answer.Sub != "" {
		return &answer, nil
	} else {
		return &answer, errors.New("WTF HUH")
	}
}

type User struct {
	Sub        string `json:"sub"`
	Name       string `json:"name"`
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
	Picture    string `json:"picture"`
	Locale     string `json:"locale"`
}
