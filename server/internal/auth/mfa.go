// mfa.go: Uses TOTP to generate QR codes/secrets and verify codes.
package gosightauth

import (
	"fmt"
	"net/http"

	"github.com/pquerna/otp/totp"
)

// TOTP MFA
func GenerateTOTPSecret(email string) (string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "GoSight",
		AccountName: email,
	})
	if err != nil {
		return "", err
	}
	return key.Secret(), nil
}

func ValidateTOTP(secret, code string) bool {
	return totp.Validate(code, secret)
}

func LoadPendingMFA(r *http.Request) (string, error) {
	cookie, err := r.Cookie("pending_mfa")
	if err != nil {
		return "", fmt.Errorf("pending_mfa cookie not found")
	}

	userID := cookie.Value
	if userID == "" {
		return "", fmt.Errorf("pending_mfa cookie was empty")
	}

	return userID, nil
}

func SavePendingMFA(userID string, w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "pending_mfa",
		Value:    userID,
		Path:     "/",
		MaxAge:   300, // 5 minutes
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}
