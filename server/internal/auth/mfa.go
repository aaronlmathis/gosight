// mfa.go: Uses TOTP to generate QR codes/secrets and verify codes.
package gosightauth

import (
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

// -- mfa cookie helpers --
func SavePendingMFA(userID string, w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "pending_mfa",
		Value:    userID,
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
		MaxAge:   300, // 5 mins
	})
}

func LoadPendingMFA(r *http.Request) (string, error) {
	c, err := r.Cookie("pending_mfa")
	if err != nil {
		return "", err
	}
	return c.Value, nil
}
