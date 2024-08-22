package utils

import (
	"strconv"

	"github.com/algrvvv/monlog/internal/config"
)

func ValidateServerId(serverId string) int {
	intID, err := strconv.Atoi(serverId)
	if err != nil {
		return -1
	}

	servers := config.Cfg.Servers
	if intID < 0 || len(servers) <= intID {
		return -1
	}

	return intID
}
