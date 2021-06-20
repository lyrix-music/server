package lastfm

import (
	"github.com/jinzhu/gorm"
	"github.com/shkh/lastfm-go/lastfm"
	"github.com/srevinsaju/lyrix/backend/config"
	"github.com/srevinsaju/lyrix/backend/types"
	"github.com/withmandala/go-log"
	"os"
)


var logger = log.New(os.Stdout)


func UpdateNowPlaying(ctx *types.Context, song types.SongMeta, userId float64) {
	if ! ServerSupportsLastFm(ctx.Config) {
		return
	}

	lastfmApi := Login(ctx, userId)
	if lastfmApi == nil {
		return
	}

	resp1, err := lastfmApi.Track.UpdateNowPlaying(map[string]interface{}{
		"artist": song.GetFirstArtist(),
		"track":  song.Track,
	})
	logger.Infof("Received response from last.fm, %s", resp1)
	if err != nil {
		logger.Warn(err)
	}

}

func Scrobble(ctx *types.Context, song types.CurrentListeningSongLocal, userId float64) {
	if ! ServerSupportsLastFm(ctx.Config) {
		return
	}

	lastfmApi := Login(ctx, userId)
	if lastfmApi == nil {
		return
	}

	resp1, err := lastfmApi.Track.Scrobble(map[string]interface{}{
		"artist": song.GetFirstArtist(),
		"track":  song.Track,
		"timestamp": song.UpdatedAt.Unix(),
	})
	logger.Infof("Received response from last.fm, %s", resp1)
	if err != nil {
		logger.Warn(err)
	}

}

func StoreAuthToken(ctx *types.Context, username string, userId float64, token string) {
	tr := ctx.Database.Model(&types.LastFmAuthToken{}).Where("id = ?", userId).Update("token", token)
	if gorm.IsRecordNotFoundError(tr.Error) || tr.RowsAffected == 0{
		// always handle error like this, cause errors maybe happened when connection failed or something.
		// record not found...

		ctx.Database.Create(&types.LastFmAuthToken{Id: int(userId), Username: username, Token: token}) // create new record from newUser

	}

	return
}

func StoreSessionKey(ctx *types.Context, username string, userId float64, sk string) {
	tr := ctx.Database.Model(&types.LastFmSessionKey{}).Where("id = ?", userId).Update("session_key", sk)
	if gorm.IsRecordNotFoundError(tr.Error) || tr.RowsAffected == 0{
		// always handle error like this, cause errors maybe happened when connection failed or something.
		// record not found...

		ctx.Database.Create(&types.LastFmSessionKey{Id: int(userId), Username: username, SessionKey: sk}) // create new record from newUser

	}

	return
}


func ServerSupportsLastFm(cfg config.Config) bool {
	return cfg.Services.LastFm.ApiKey != "" && cfg.Services.LastFm.SharedSecret != ""
}


func New(cfg config.Config) *lastfm.Api {
	if ServerSupportsLastFm(cfg) {
		return lastfm.New(cfg.Services.LastFm.ApiKey, cfg.Services.LastFm.SharedSecret)
	}
	return nil
}


func Login(ctx *types.Context, userId float64) *lastfm.Api {
	if !ServerSupportsLastFm(ctx.Config) {
		return nil
	}
	lastfmApi := New(ctx.Config)
	lastFmAuthToken := types.LastFmAuthToken{}
	respToken := ctx.Database.First(&lastFmAuthToken, "id = ?", userId)
	if respToken.Error != nil || respToken.RowsAffected == 0 {
		return nil
	}
	// if the user does not exist on the token database, but empty token
	if lastFmAuthToken.Token == "" {
		return nil
	}

	lastFmSessionKey := types.LastFmSessionKey{}
	respSk := ctx.Database.First(&lastFmSessionKey, "id = ?", userId)
	if respSk.Error != nil || respSk.RowsAffected == 0 {
		// the user has a token which is not used yet
		// need to create a session key and use it now
		err := lastfmApi.LoginWithToken(lastFmAuthToken.Token)
		if err != nil {
			logger.Warn(err)
			return nil
		}
		StoreSessionKey(ctx, "", userId, lastfmApi.GetSessionKey())
	} else {
		lastfmApi.SetSession(lastFmSessionKey.SessionKey)
	}
	return lastfmApi
}