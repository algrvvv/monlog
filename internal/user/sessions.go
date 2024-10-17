package user

var (
	// sessions список сессий, который используется только, если в конфиге включена опция config.Cfg.App.Auth = true
	sessions = make(map[string]string)
)

// SetSession функция для добавления нового токена в активные сессии
func SetSession(token, username string) {
	sessions[token] = username
}

// GetSession функция для получения юзер нейма по токену сессии
func GetSession(token string) string {
	return sessions[token]
}

func DeleteSession(token string) {
	delete(sessions, token)
}

func SessionExists(token string) bool {
	_, ok := sessions[token]
	return ok
}
