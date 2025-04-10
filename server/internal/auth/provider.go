package gosightauth

import (
	"net/http"

	"github.com/aaronlmathis/gosight/server/internal/usermodel"
)

// provider.go: Defines AuthProvider interface that local.go and sso/*.go implement.
// AuthProvider is implemented by all authentication types
// including local login and SSO providers like Google.

type AuthProvider interface {
	StartLogin(w http.ResponseWriter, r *http.Request)
	HandleCallback(w http.ResponseWriter, r *http.Request) (*usermodel.User, error)
}
