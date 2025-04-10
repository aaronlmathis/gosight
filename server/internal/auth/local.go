// local.go: Handles login, registration, password validation.
package gosightauth

import (
	"net/http"

	"github.com/aaronlmathis/gosight/server/internal/store/userstore"
	"github.com/aaronlmathis/gosight/server/internal/usermodel"
	"github.com/aaronlmathis/gosight/shared/utils"
)

type LocalAuth struct {
	Store userstore.UserStore
}

func (l *LocalAuth) StartLogin(w http.ResponseWriter, r *http.Request) {
	// Local login doesn't require redirect — handled via UI
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Local login requested. UI should render login form."))
}

func (l *LocalAuth) HandleCallback(w http.ResponseWriter, r *http.Request) (*usermodel.User, error) {
	ctx := r.Context()
	username := r.FormValue("username")
	password := r.FormValue("password")
	totp := r.FormValue("totp")

	user, err := l.Store.GetUserByUsername(ctx, username)
	if err != nil {
		utils.Debug("❌ User not found: %s", username)
		return nil, err
	}
	if !CheckPasswordHash(password, user.PasswordHash) {
		utils.Debug("❌ Password Hash ain't right: %")
		return nil, ErrInvalidPassword
	}
	if user.TOTPSecret != "" && !ValidateTOTP(user.TOTPSecret, totp) {
		return nil, ErrInvalidTOTP
	}
	return user, nil
}
