package main

import (
	"fmt"
	"github.com/lyrix-music/server/meta"
	lxlfm "github.com/lyrix-music/server/services/lastfm"
	"os"
	"strings"

	"github.com/withmandala/go-log"

	"github.com/lyrix-music/server/config"
	"github.com/lyrix-music/server/database"
	"github.com/lyrix-music/server/routes"
	"github.com/lyrix-music/server/types"
)

var logger = log.New(os.Stdout)

func main() {

	command := os.Args[len(os.Args)-1]
	logger.Infof("%s Build:%s %s", meta.AppName, meta.BuildVersion, meta.BuildTime)

	if !strings.HasSuffix(command, ".json") && !strings.HasSuffix(command, ".yaml") {
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

	lastFmApi := lxlfm.New(cfg)
	// create a context
	ctx := &types.Context{Database: db, Config: cfg, LastFm: lastFmApi}

	// create a http rest api instance
	app, err := routes.Initialize(cfg, ctx)
	if err != nil {
		logger.Fatal(err)
	}
	app.Listen(fmt.Sprintf(":%d", cfg.Server.Port))

}
