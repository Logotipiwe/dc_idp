package main

import (
	"github.com/logotipiwe/dc_go_utils/src/config"
	"net/http"
)

func setRedirectAfterAuthCookie(w http.ResponseWriter, destination string) {
	cookie := http.Cookie{
		Name:     redirAfterAuthCookieName,
		Value:    destination,
		HttpOnly: false,
		Path:     "/",
	}
	http.SetCookie(w, &cookie)
}

func setATCookie(w http.ResponseWriter, token string) {
	cookie := http.Cookie{
		Name:     "access_token",
		Value:    token,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, &cookie)
}

func getFallbackRedirect() string {
	return config.GetConfig("FALLBACK_REDIRECT")
}

func isGoogleAutoAuth() bool {
	return config.GetConfig("GOOGLE_AUTO_AUTH") == "1"
}
