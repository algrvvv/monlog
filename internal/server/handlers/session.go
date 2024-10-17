package handlers

import (
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/algrvvv/monlog/internal/config"
	"github.com/algrvvv/monlog/internal/logger"
	"github.com/algrvvv/monlog/internal/user"
	"github.com/algrvvv/monlog/internal/utils"
)

func Login(w http.ResponseWriter, r *http.Request) {
	formUsername := r.PostFormValue("username")
	formPass := r.PostFormValue("password")

	corrUsername, corrPass := user.GetUserData()
	isPassValid := user.CheckPassword(formPass, corrPass)

	if formUsername != corrUsername || !isPassValid {
		logger.Warnf("failed login account attempt by user: %s", formUsername)
		utils.RenderError(w, "Incorrect login or password", http.StatusForbidden)
		return
	}
	logger.Info("successful login account")

	sessionToken := uuid.NewString()
	cookie := http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   config.Cfg.App.Debug,
	}
	http.SetCookie(w, &cookie)
	user.SetSession(sessionToken, formUsername)

	http.Redirect(w, r, "/", http.StatusFound)
}
