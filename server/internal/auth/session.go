package gosightauth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// session.go: Handles signed cookies or token generation/validation.
var jwtSecret = []byte("your-super-secret-key-change-this")

type SessionClaims struct {
	UserID           string   `json:"sub"`
	Roles            []string `json:"roles,omitempty"`
	TraceID          string   `json:"trace_id,omitempty"`
	RolesRefreshedAt int64    `json:"roles_refreshed_at"`
	jwt.RegisteredClaims
}

func GenerateToken(userID string, roles []string, traceID string) (string, error) {
	claims := SessionClaims{
		UserID:           userID,
		Roles:            roles,
		TraceID:          traceID,
		RolesRefreshedAt: time.Now().Unix(),
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(2 * time.Hour)),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ValidateToken(tokenStr string) (*SessionClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &SessionClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*SessionClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

func SetSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "gosight_session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(2 * time.Hour),
	})
}

var ErrNoSession = errors.New("no session token found")

// GetSessionToken retrieves the session token from cookie or header
func GetSessionToken(r *http.Request) (string, error) {
	if cookie, err := r.Cookie("gosight_session"); err == nil && cookie.Value != "" {
		return cookie.Value, nil
	}

	authHeader := r.Header.Get("Authorization")
	if len(authHeader) > 7 && strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer "), nil
	}

	return "", ErrNoSession
}

// GetSessionClaims retrieves the session claims from the request
func GetSessionClaims(r *http.Request) (*SessionClaims, error) {
	token, err := GetSessionToken(r)
	if err != nil {
		return nil, err
	}
	return ValidateToken(token)
}

// Convenience: get user ID from session token in request
func GetSessionUserID(r *http.Request) (string, error) {
	claims, err := GetSessionClaims(r)
	if err != nil {
		return "", err
	}
	return claims.UserID, nil
}
