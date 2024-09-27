package utils

import (
	"strconv"

	"github.com/algrvvv/monlog/internal/config"
)

func ValidateServerID(serverID string) int {
	intID, err := strconv.Atoi(serverID)
	if err != nil {
		return -1
	}

	servers := config.Cfg.Servers
	if intID < 0 || len(servers) <= intID {
		return -1
	}

	return intID
}
