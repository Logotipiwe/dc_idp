package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Logotipiwe/dc_go_auth_lib/auth"
	env "github.com/logotipiwe/dc_go_env_lib"
	"html/template"
	"net/http"
	"net/url"
	"os"
)

const clientId = "319710408255-ntkf14k8ruk4p98sn2u1ho4j99rpjqja.apps.googleusercontent.com"
const redirAfterAuthCookieName = "redirect_after_auth"

func main() {
	err := InitDb()
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	})

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
		token := exchangeGoogleCodeToToken(code)
		println("Token is: " + token)
		gUser, err := getGoogleUserDataFromGoogleAT(token)
		user := createDcUserFromGoogleUser(gUser)
		exists, err := existsInDbByGoogleId(gUser.Sub)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(500) //TODO ??
			return
		}
		if exists {
			fmt.Println("Dc user exists in db, receiving it...")
			user, err = getUserFromDbByGoogleId(gUser.Sub)
			if err != nil {
				fmt.Println(err)
				w.WriteHeader(500)
				return
			}
		} else {
			fmt.Println("Dc user doesn't exist. Creating...")
			user, err = createUserInDb(user)
			if err != nil {
				fmt.Println(err)
				w.WriteHeader(500)
				return
			}
		}
		fmt.Printf("Dc user is : %v\n", user)

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
		println("/getUser")
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3003") //TODO only on dev
		w.Header().Set("Access-Control-Allow-Credentials", "true")             //TODO only on dev
		user, err := getUserData(r)
		if err != nil {
			println("Error getting user: %s", err.Error())
			w.WriteHeader(403)
			return
		}
		marshal, err := json.Marshal(&user)
		if err != nil {
			w.WriteHeader(500)
			return
		}
		fmt.Fprint(w, string(marshal))
	})

	http.HandleFunc("/get-user-by-id", func(w http.ResponseWriter, r *http.Request) {
		println("/get-user-by-id")
		w.Header().Set("Access-Control-Allow-Origin", "*") //TODO only on dev

		err = auth.AuthAsMachine(r)
		if err != nil {
			println("Error: Cannot auth as machine", err.Error())
			w.WriteHeader(401)
			return
		}
		userID := r.URL.Query().Get("userId")
		user, err := getUserFromDbById(userID)
		if err != nil {
			println("Error getting user: %s", err.Error())
			w.WriteHeader(500)
			return
		}
		marshal, err := json.Marshal(&user)
		if err != nil {
			w.WriteHeader(500)
			return
		}
		fmt.Fprint(w, string(marshal))
	})

	println("Ready")
	err = http.ListenAndServe(":"+os.Getenv("CONTAINER_PORT"), nil)
	println("Server up!")
	if err != nil {
		panic("Lol server fell")
	}
}
