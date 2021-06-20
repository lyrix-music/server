package lastfm

import (
	"github.com/jinzhu/gorm"
	"github.com/srevinsaju/lyrix/backend/types"
	"github.com/withmandala/go-log"
	"os"
)


var logger = log.New(os.Stdout)


func UpdateNowPlaying(ctx *types.Context, song types.SongMeta, userId float64) {
	defer ctx.LastFm.SetSession("")

	if ctx.LastFm == nil {
		return
	}

	userInDatabase := types.LastFmAuthToken{}
	resp := ctx.Database.First(&userInDatabase, "id = ?", userId)
	// if the user does not exist on the
	if resp.Error != nil || resp.RowsAffected == 0 {
		return
	}
	if userInDatabase.Token == "" {
		return
	}
	logger.Info(userInDatabase)

	err := ctx.LastFm.LoginWithToken(userInDatabase.Token)
	if err != nil {
		logger.Warn(err)
		// restore to logged-out state
		return
	}

	resp1, err := ctx.LastFm.Track.UpdateNowPlaying(map[string]interface{}{
		"artist": song.GetFirstArtist(),
		"track":  song.Track,
	})
	logger.Infof("Received response from last.fm, %s", resp1)
	if err != nil {
		logger.Warn(err)
	}

}

func StoreAuthToken(ctx *types.Context, username string, userId float64, token string) {
	tr := ctx.Database.Model(&types.LastFmAuthToken{}).Where("id = ?", username).Update("token", token)
	if gorm.IsRecordNotFoundError(tr.Error) || tr.RowsAffected == 0{
		// always handle error like this, cause errors maybe happened when connection failed or something.
		// record not found...

		ctx.Database.Create(&types.LastFmAuthToken{Id: int(userId), Username: username, Token: token}) // create new record from newUser

	}

	return
}