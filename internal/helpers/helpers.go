package helpers

import (
	"errors"
	"github.com/lyrix-music/server/types"
	"strings"
)


var UserIdParseError = errors.New("the user id is not fully qualified, eg: @something@somewhere.xyz")
// var logger = log.New(os.Stdout)

func ParseFullUserId(userId string) (types.User, error) {
	parts := strings.Split(userId, "@")
	// logger.Info(parts)
	if len(parts) != 3 {
		return types.User{}, UserIdParseError
	}
	return types.User{Username: parts[1], Homeserver: parts[2]}, nil
}