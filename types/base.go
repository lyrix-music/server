package types

import (
	"github.com/jinzhu/gorm"
	"github.com/shkh/lastfm-go/lastfm"
	"github.com/lyrix-music/server/config"
)



type Context struct {
	Database *gorm.DB
	Config config.Config
	LastFm *lastfm.Api
}


