package middlewares

import (
	"errors"
	"net/http"
	"strings"

	"github.com/algrvvv/monlog/internal/config"
	"github.com/algrvvv/monlog/internal/logger"
	"github.com/algrvvv/monlog/internal/user"
	"github.com/algrvvv/monlog/internal/utils"
)

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !config.Cfg.App.Auth {
			next.ServeHTTP(w, r)
			return // я рот ебал этого забытого ретерна нахуй
		}

		if r.URL.Path == "/login" && (r.Method == "POST" || r.Method == "GET") ||
			strings.HasPrefix(r.URL.Path, "/static/") {
			next.ServeHTTP(w, r)
			return
		}

		cookie, err := r.Cookie("session_token")
		if err != nil {
			logger.Warnf("failed to read cookies: %v", err)
			if errors.Is(err, http.ErrNoCookie) {
				logger.Info("redirect user to login page")
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}
			logger.Errorf("failed to read cookies: %v", err)
			utils.RenderError(w, "failed to read cookies", http.StatusInternalServerError)
			return
		}

		if user.SessionExists(cookie.Value) {
			next.ServeHTTP(w, r)
			return // и этого тоже D:
		}

		http.Redirect(w, r, "/login", http.StatusFound)
	})
}
