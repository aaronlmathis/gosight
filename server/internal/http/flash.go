// internal/http/flash.go
// Sets cookie based flash messages

package httpserver

import (
	"net/http"
	"net/url"
	"time"
)

const flashCookieName = "gosight_flash"

func SetFlash(w http.ResponseWriter, message string) {
	http.SetCookie(w, &http.Cookie{
		Name:     flashCookieName,
		Value:    url.QueryEscape(message),
		Path:     "/",
		MaxAge:   10,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

func GetFlash(w http.ResponseWriter, r *http.Request) string {
	cookie, err := r.Cookie(flashCookieName)
	if err != nil {
		return ""
	}

	// Clear it immediately
	http.SetCookie(w, &http.Cookie{
		Name:     flashCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	value, _ := url.QueryUnescape(cookie.Value)
	return value
}
