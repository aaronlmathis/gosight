// mfa.go: Uses TOTP to generate QR codes/secrets and verify codes.
package gosightauth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aaronlmathis/gosight/shared/utils"
	"github.com/pquerna/otp/totp"
)

var mfaKey []byte

func InitMFAKey(encoded string) error {
	key, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(key) != 32 {
		return fmt.Errorf("invalid MFA key: must be 32 bytes after base64 decoding")
	}
	mfaKey = key
	return nil
}

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

// LoadPendingMFA retrieves the pending MFA cookie
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

// SavePendingMFA sets a cookie to remember the pending MFA for 5 minutes
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

// SetRememberMFA sets a cookie to remember the MFA for 30 days - bind it to device
func SetRememberMFA(w http.ResponseWriter, userID string, r *http.Request) {
	expiry := time.Now().Add(30 * 24 * time.Hour).Unix()
	fingerprint := hashUserAgent(r.UserAgent())
	plaintext := fmt.Sprintf("%s|%d|%s", userID, expiry, fingerprint)

	block, err := aes.NewCipher(mfaKey)
	if err != nil {
		utils.Error("MFA cipher error: %v", err)
		return
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		utils.Error("MFA GCM error: %v", err)
		return
	}

	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		utils.Error("MFA nonce error: %v", err)
		return
	}

	ciphertext := aesgcm.Seal(nil, nonce, []byte(plaintext), nil)
	final := append(nonce, ciphertext...) // prepend nonce for decryption

	http.SetCookie(w, &http.Cookie{
		Name:     "remember_mfa",
		Value:    base64.URLEncoding.EncodeToString(final),
		Path:     "/",
		Expires:  time.Unix(expiry, 0),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

func CheckRememberMFA(r *http.Request, userID string) bool {

	cookie, err := r.Cookie("remember_mfa")
	if err != nil {
		return false
	}

	data, err := base64.URLEncoding.DecodeString(cookie.Value)
	if err != nil {
		return false
	}

	block, err := aes.NewCipher(mfaKey)
	if err != nil {
		return false
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return false
	}

	if len(data) < aesgcm.NonceSize() {
		return false
	}

	nonce := data[:aesgcm.NonceSize()]
	ciphertext := data[aesgcm.NonceSize():]

	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return false
	}

	parts := strings.Split(string(plaintext), "|")
	if len(parts) != 3 {
		return false
	}

	fingerprint := hashUserAgent(r.UserAgent())
	if parts[2] != fingerprint {
		return false
	}

	id := parts[0]
	expiryUnix, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || time.Now().Unix() > expiryUnix {
		return false
	}

	return id == userID
}

func ClearRememberMFA(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "remember_mfa",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

// hashUserAgent hashes the user agent string using SHA-256
func hashUserAgent(ua string) string {
	sum := sha256.Sum256([]byte(ua))
	return hex.EncodeToString(sum[:])
}
