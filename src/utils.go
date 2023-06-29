package main

import (
	"fmt"
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

func getFallbackRedirect() string {
	return os.Getenv("FALLBACK_REDIRECT")
}
