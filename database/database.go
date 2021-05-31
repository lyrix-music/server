package database

import (
	"os"

	"github.com/withmandala/go-log"

	// Import GORM-related packages.
	// "github.com/cockroachdb/cockroach-go/crdb/crdbgorm"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"

	"github.com/srevinsaju/lyrix/backend/config"
	"github.com/srevinsaju/lyrix/backend/types"
)

var logger = log.New(os.Stdout)

func Connect(cfg config.Config) (*gorm.DB, error) {

	addr := cfg.Backend.ConnectionString
	db, err := gorm.Open("postgres", addr)
	if err != nil {
		logger.Fatal(err)
	}

	// Set to `true` and GORM will print out all DB queries.
	db.LogMode(cfg.Backend.Debug)

	// Automatically create the tables
	db.AutoMigrate(&types.UserAccount{})
	db.AutoMigrate(&types.CurrentListeningSongLocal{})

	db.AutoMigrate(&types.SpotifyAuthToken{})

	return db, nil

}
