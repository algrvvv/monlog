package middlewares

import (
	"errors"
	"net/http"

	"github.com/algrvvv/monlog/internal/config"
	"github.com/algrvvv/monlog/internal/logger"
	"github.com/algrvvv/monlog/internal/user"
)

func Guest(r *http.Request) bool {
	if !config.Cfg.App.Auth {
		return false
	}

	cookie, err := r.Cookie("session_token")
	if err != nil {
		logger.Warnf("failed to read cookies: %v", err)
		if errors.Is(err, http.ErrNoCookie) {
			logger.Info("redirect user to login page")
		} else {
			logger.Errorf("failed to read cookies: %v", err)
		}
		return true
	}

	return !user.SessionExists(cookie.Value)
}
