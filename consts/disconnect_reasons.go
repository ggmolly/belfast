package consts

import "fmt"

const (
	DR_LOGGED_IN_ON_ANOTHER_DEVICE = 1
	DR_SERVER_MAINTENANCE          = 2
	DR_GAME_UPDATE                 = 3
	DR_OFFLINE_TOO_LONG            = 4
	DR_CONNECTION_LOST             = 5
	DR_CONNECTION_TO_SERVER_LOST   = 6
	DR_DATA_VALIDATION_FAILED      = 7
	DR_LOGIN_DATA_EXPIRED          = 199
)

func ResolveReason(reason uint8) string {
	switch reason {
	case DR_LOGGED_IN_ON_ANOTHER_DEVICE:
		return "logged in on another device"
	case DR_SERVER_MAINTENANCE:
		return "server maintenance"
	case DR_GAME_UPDATE:
		return "game update"
	case DR_OFFLINE_TOO_LONG:
		return "offline too long"
	case DR_CONNECTION_LOST:
		return "connection lost"
	case DR_CONNECTION_TO_SERVER_LOST:
		return "connection to server lost"
	case DR_DATA_VALIDATION_FAILED:
		return "data validation failed"
	case DR_LOGIN_DATA_EXPIRED:
		return "login data expired"
	default:
		return fmt.Sprintf("unknown reason %d", reason)
	}
}
