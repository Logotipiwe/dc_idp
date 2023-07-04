package main

import (
	"net/http"
	"os"
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
	return os.Getenv("FALLBACK_REDIRECT")
}
