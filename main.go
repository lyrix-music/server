package main

import (
	"fmt"
	"github.com/shkh/lastfm-go/lastfm"
	"os"
	"strings"

	"github.com/withmandala/go-log"

	"github.com/srevinsaju/lyrix/backend/config"
	"github.com/srevinsaju/lyrix/backend/database"
	"github.com/srevinsaju/lyrix/backend/routes"
	"github.com/srevinsaju/lyrix/backend/types"
)

var logger = log.New(os.Stdout)

const (
	BuildName    = "Lyrix Backend"
	BuildVersion = "(local dev build)"
	BuildTime    = ""
)

func main() {

	command := os.Args[len(os.Args)-1]
	logger.Infof("%s Build:%s %s", BuildName, BuildVersion, BuildTime)

	if !strings.HasSuffix(command, ".json") {
		// the user has not provided any commands along with the executable name
		// so, we should show the usage
		logger.Info("To load an existing configuration: ")
		logger.Info("  $ ./backend path/to/config.json")
		return
	}

	if _, err := os.Stat(command); os.IsNotExist(err) {
		logger.Fatal("The specified path does not exist")
	}

	// get the path configuration and read the configuration
	configPath := command
	cfg := config.ParseFromFile(configPath)

	// initialize the connection to the database
	db, err := database.Connect(cfg)
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()

	var lastFmApi *lastfm.Api
	if cfg.Services.LastFm.ApiKey != "" && cfg.Services.LastFm.SharedSecret != "" {
		lastFmApi = lastfm.New(cfg.Services.LastFm.ApiKey, cfg.Services.LastFm.SharedSecret)
		logger.Info("Last.fm support has been enabled.")
		logger.Infof("last.fm authorize url: %s",
			lastFmApi.GetAuthRequestUrl(fmt.Sprintf("%s/callback/lastfm/token", cfg.Server.PublicEndpoint)),
		)
	}
	// create a context
	ctx := &types.Context{Database: db, Config: cfg, LastFm: lastFmApi}

	// create a http rest api instance
	app, err := routes.Initialize(cfg, ctx)
	if err != nil {logger.Fatal(err)}
	app.Listen(fmt.Sprintf(":%d", cfg.Server.Port))

}
