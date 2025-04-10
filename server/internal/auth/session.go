package gosightauth

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"time"
)

// session.go: Handles signed cookies or token generation/validation.
var hmacSecret = []byte("replace-with-secure-key")

func GenerateToken(userID string) string {
	h := hmac.New(sha256.New, hmacSecret)
	h.Write([]byte(userID))
	sig := h.Sum(nil)
	return base64.StdEncoding.EncodeToString([]byte(userID + "." + base64.StdEncoding.EncodeToString(sig)))
}

func ParseToken(token string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return "", err
	}
	parts := bytes.Split(data, []byte("."))

	if len(parts) != 2 {
		return "", errors.New("invalid token")
	}
	userID := string(parts[0])
	h := hmac.New(sha256.New, hmacSecret)
	h.Write([]byte(userID))
	if !hmac.Equal(h.Sum(nil), mustDecodeBase64(string(parts[1]))) {
		return "", errors.New("invalid signature")
	}
	return userID, nil
}

func mustDecodeBase64(s string) []byte {
	b, _ := base64.StdEncoding.DecodeString(s)
	return b
}

// Cookie Helpers
func SetSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(24 * time.Hour),
	})
}

func GetSessionUserID(r *http.Request) (string, error) {
	cookie, err := r.Cookie("session")
	if err != nil {
		return "", err
	}
	return ParseToken(cookie.Value)
}
