package types

import (
	"github.com/jinzhu/gorm"
	"github.com/shkh/lastfm-go/lastfm"
	"github.com/srevinsaju/lyrix/backend/config"
)



type Context struct {
	Database *gorm.DB
	Config config.Config
	LastFm *lastfm.Api
}


