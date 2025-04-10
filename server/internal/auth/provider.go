package gosightauth

import (
	"net/http"

	"github.com/aaronlmathis/gosight/server/internal/store/userstore"
)

// provider.go: Defines AuthProvider interface that local.go and sso/*.go implement.

type AuthProvider interface {
	StartLogin(w http.ResponseWriter, r *http.Request)
	HandleCallback(w http.ResponseWriter, r *http.Request) (*userstore.User, error)
}
