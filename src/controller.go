package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	env "github.com/logotipiwe/dc_go_env_lib"
	"html/template"
	"net/http"
	"net/url"
	"os"
)

const clientId = "319710408255-ntkf14k8ruk4p98sn2u1ho4j99rpjqja.apps.googleusercontent.com"
const redirAfterAuthCookieName = "redirect_after_auth"

func main() {
	http.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		println("/auth")
		redirect := r.URL.Query().Get("redirect")
		googleRedirect := env.GetCurrUrl() + env.GetSubpath() + "/g-oauth"
		googleOauthUri := "https://accounts.google.com/o/oauth2/v2/auth?" +
			"client_id=" + url.QueryEscape(clientId) +
			"&redirect_uri=" + url.QueryEscape(googleRedirect) +
			"&response_type=code&scope=profile"
		println(fmt.Sprintf("Request will be redirected to : %v", redirect))
		setRedirectAfterAuthCookie(w, redirect)
		http.Redirect(w, r, googleOauthUri, 302)
	})

	http.HandleFunc("/g-oauth", func(w http.ResponseWriter, r *http.Request) {
		println("/g-oauth")
		code := r.URL.Query().Get("code")
		println("Code is: " + code)
		token := exchangeCodeToToken(code)
		println("Token is: " + token)
		setATCookie(w, token)
		redirTo := getRedirectUri(r)
		println("Redirecting after auth: " + redirTo)
		http.Redirect(w, r, redirTo, 302)
	})

	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		setATCookie(w, "")
		redirTo := r.URL.Query().Get("redirect")
		if redirTo == "" {
			redirTo = getFallbackRedirect()
		}
		http.Redirect(w, r, redirTo, 302)
	})

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("src/templates/login.gohtml"))
		redirTo := r.URL.Query().Get("redirect")
		var tpl bytes.Buffer
		authHref := env.GetSubpath() + "/auth?redirect=" + url.QueryEscape(redirTo)
		if err := tmpl.Execute(&tpl, authHref); err != nil {
			fmt.Fprintf(w, "Internal error")
		}

		fmt.Fprint(w, tpl.String())
	})

	http.HandleFunc("/getUser", func(w http.ResponseWriter, r *http.Request) {
		user, err := getUserData(r)
		if err != nil {
			w.WriteHeader(403)
			return
		}
		marshal, err := json.Marshal(user)
		if err != nil {
			w.WriteHeader(403)
			return
		}
		fmt.Fprint(w, string(marshal))
	})

	err := http.ListenAndServe(":"+os.Getenv("CONTAINER_PORT"), nil)
	println("Server up!")
	if err != nil {
		panic("Lol server fell")
	}
}
