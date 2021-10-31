package ops

import (
	"github.com/lyrix-music/server/types"
)

func GetCurrentSongForUserId(ctx *types.Context, userId int) types.CurrentListeningSongLocal {
	userInDatabase := types.CurrentListeningSongLocal{}
	resp := ctx.Database.First(&userInDatabase, "id = ?", userId)
	if resp.Error != nil {
		logger.Warn(resp.Error)
		return types.CurrentListeningSongLocal{}
	}
	return userInDatabase
}
