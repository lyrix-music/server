package types

import (
	"github.com/jinzhu/gorm"
	"github.com/lyrix-music/server/config"
	"github.com/shkh/lastfm-go/lastfm"
)

type Context struct {
	Database *gorm.DB
	Config   config.Config
	LastFm   *lastfm.Api
}
