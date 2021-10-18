package routes

import (
	"encoding/json"
	"errors"
	"github.com/lyrix-music/server/internal/helpers"
	"github.com/lyrix-music/server/meta"
	"github.com/lyrix-music/server/services/lastfm"
	"os"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/lyrix-music/server/config"
	"github.com/lyrix-music/server/types"
	"github.com/withmandala/go-log"
	"golang.org/x/crypto/bcrypt"

	sl "github.com/srevinsaju/swaglyrics-go"
	slTypes "github.com/srevinsaju/swaglyrics-go/types"

	jwt "github.com/form3tech-oss/jwt-go"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	mwLogger "github.com/gofiber/fiber/v2/middleware/logger"
	jwtware "github.com/gofiber/jwt/v2"
)

var logger = log.New(os.Stdout)

func Initialize(cfg config.Config, ctx *types.Context) (*fiber.App, error) {

	app := fiber.New(fiber.Config{DisableStartupMessage: true})

	logger.Infof("Preparing listener")

	isUserExists := func(username string) (bool, error) {
		if username == "" {
			return false, errors.New("no username provided")
		}

		userInDatabase := types.UserAccount{}
		ctx.Database.Where("username = ?", username).Find(&userInDatabase)
		if userInDatabase.Username == username {
			return true, nil
		}
		return false, nil
	}

	app.Use(cors.New())
	limits := limiter.New(limiter.Config{
		Next: func(c *fiber.Ctx) bool {
			return c.IP() == "127.0.0.1"
		},
		Max:        20,
		Expiration: 30 * time.Second,
		KeyGenerator: func(c *fiber.Ctx) string {
			id := c.Get("X-Forwarded-for")
			if id == "" {
				return c.IP()
			}
			return id
		},
	})

	// Unauthenticated route
	app.Get("/", accessible)

	app.Use(mwLogger.New())

	app.Get("/version", func(c *fiber.Ctx) error {
		return c.SendString(meta.BuildVersion)
	})

	app.Use("/register", limits)
	// Register
	app.Post("/register", func(c *fiber.Ctx) error {
		// data:UserAccountRegister
		user := &types.UserAccountRegister{}

		err := json.Unmarshal(c.Body(), user)
		if err != nil {
			return err
		}
		if user.Username == "" {
			return c.SendStatus(fiber.StatusUnauthorized)
		}
		isUserAlreadyExists, err := isUserExists(user.Username)
		if err != nil {
			return err
		}
		if isUserAlreadyExists {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		count := -1
		ctx.Database.Model(&types.UserAccount{}).Count(&count)

		userHashed, err := user.Hash()
		if err != nil {
			logger.Warn(err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		userHashed.Id = count + 1
		logger.Infof("Registration from user %s", user.Username)
		if user.Username == "" || user.Password == "" || user.TelegramId == 0 {
			return c.SendStatus(fiber.StatusUnauthorized)
		}
		ctx.Database.Create(&userHashed)
		return c.SendStatus(fiber.StatusAccepted)

	})
	app.Use("/login", limits)
	// Login route
	app.Post("/login", func(c *fiber.Ctx) error {
		// data:UserAccountRegister
		user := &types.UserAccountRegister{}

		err := json.Unmarshal(c.Body(), user)
		if err != nil {
			return err
		}
		if user.Username == "" {
			return c.SendStatus(fiber.StatusUnauthorized)
		}

		// check if the user is registered in the first place.
		userInDatabase := types.UserAccount{}
		ctx.Database.Where("username = ?", user.Username).Find(&userInDatabase)
		if userInDatabase.Username == "" {
			return c.SendStatus(fiber.StatusUnauthorized)
		}

		// the password doesnt match
		err = bcrypt.CompareHashAndPassword([]byte(userInDatabase.HashedPassword), []byte(user.Password))
		if err != nil {
			return c.SendStatus(fiber.StatusUnauthorized)
		}

		return loginJWT(c, userInDatabase, cfg.SecretKey)

	})

	app.Get("/user/exists", func(c *fiber.Ctx) error {
		// data:UserAccountRegister
		user := &types.UserAccountRegister{}

		err := json.Unmarshal(c.Body(), user)
		if err != nil {
			return err
		}
		if user.Username == "" {
			return c.SendStatus(fiber.StatusUnauthorized)
		}
		isUserAlreadyExists, err := isUserExists(user.Username)
		if err != nil {
			return err
		}
		if isUserAlreadyExists {
			return c.JSON(map[string]string{"exists": strconv.FormatBool(isUserAlreadyExists)})
		}
		return c.JSON(map[string]string{"exists": "yes"})

	})
	// JWT Middleware

	app.Use("/user", jwtware.New(jwtware.Config{
		SigningKey: []byte(cfg.SecretKey),
	}))
	app.Use("/dot", jwtware.New(jwtware.Config{
		SigningKey: []byte(cfg.SecretKey),
	}))

	app.Use("/connect/lastfm", jwtware.New(jwtware.Config{
		SigningKey: []byte(cfg.SecretKey),
	}))

	// Restricted Routes
	app.Get("/user/welcome", func(c *fiber.Ctx) error {
		user := c.Locals("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		name := claims["user"].(string)
		return c.SendString("Welcome " + name)
	})

	app.Get("/user/telegram_id", func(c *fiber.Ctx) error {
		user := c.Locals("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		telegramId := claims["tgid"].(string)
		return c.SendString(telegramId)
	})

	app.Get("/user/telegram_id", func(c *fiber.Ctx) error {
		user := c.Locals("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		telegramId := claims["tgid"].(string)
		return c.SendString(telegramId)
	})

	app.Get("/user/service/lastfm/token", func(c *fiber.Ctx) error {
		user := c.Locals("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		userId := claims["id"].(float64)

		userInDatabase := types.LastFmAuthToken{}
		resp := ctx.Database.First(&userInDatabase, "id = ?", userId)
		if resp.Error != nil {
			return c.JSON(types.LastFmAuthToken{})
		}
		return c.JSON(userInDatabase)
	})

	// POST /user/service/lastfm/token
	app.Post("/user/service/lastfm/token", func(c *fiber.Ctx) error {
		// data:LastFmAuthTokenRegisterRequest
		lastFmToken := &types.LastFmAuthTokenRegisterRequest{}

		err := json.Unmarshal(c.Body(), lastFmToken)
		if err != nil {
			return err
		}

		user := c.Locals("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		userId := claims["id"].(float64)
		username := claims["user"].(string)
		lastfm.StoreAuthToken(ctx, username, userId, lastFmToken.Token)
		return c.SendStatus(fiber.StatusAccepted)
	})

	// POST /user/player/spotify/token
	app.Post("/user/player/spotify/token", func(c *fiber.Ctx) error {
		// data:SpotifyAuthTokenRegisterRequest
		spotifyToken := &types.SpotifyAuthTokenRegisterRequest{}

		err := json.Unmarshal(c.Body(), spotifyToken)
		if err != nil {
			return err
		}

		user := c.Locals("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		userId := claims["id"].(float64)
		username := claims["user"].(string)
		tr := ctx.Database.Model(&types.SpotifyAuthToken{}).Where("username = ?", username).Update("token", spotifyToken.Token)
		if gorm.IsRecordNotFoundError(tr.Error) || tr.RowsAffected == 0 {
			// always handle error like this, cause errors maybe happened when connection failed or something.
			// record not found...
			ctx.Database.Create(&types.SpotifyAuthToken{Id: int(userId), Username: username, Token: spotifyToken.Token}) // create new record from newUser

		}

		return c.SendStatus(fiber.StatusAccepted)
	})

	// GET /user/player/spotify/token
	app.Get("/user/player/spotify/token", func(c *fiber.Ctx) error {
		user := c.Locals("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		userId := claims["id"].(float64)

		userInDatabase := types.SpotifyAuthToken{}
		resp := ctx.Database.First(&userInDatabase, "id = ?", userId)
		if resp.Error != nil {
			return c.JSON(types.SpotifyAuthToken{})
		}
		return c.JSON(userInDatabase)
	})

	app.Post("/user/player/local/current_song", func(c *fiber.Ctx) error {
		// data:SongMeta
		currentSong := &types.SongMeta{}

		err := json.Unmarshal(c.Body(), currentSong)
		if err != nil {
			return err
		}

		user := c.Locals("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		userId := claims["id"].(float64)
		username := claims["user"].(string)
		logger.Info("Attempting to update")
		// logger.Info(currentSong)

		if currentSong.Artist == "" || currentSong.Track == "" {

			err := ctx.Database.Model(
				&types.CurrentListeningSongLocal{}).
				Where("id = ?", userId).
				Updates(map[string]interface{}{"track": "", "artist": "", "source": "", "url": ""}).
				Error
			if err != nil {
				logger.Warn(err)
				return c.SendStatus(fiber.StatusNotFound)
			}
			return c.SendStatus(fiber.StatusAccepted)
		} else {
			if lastfm.ServerSupportsLastFm(ctx.Config) && currentSong.Scrobble {
				logger.Info("server supports lastfm, attempting scrobble")
				lastListenedSong := types.CurrentListeningSongLocal{}
				resp := ctx.Database.First(&lastListenedSong, "id = ?", userId)
				logger.Info("received last listened song", lastListenedSong)
				if resp.Error == nil && resp.RowsAffected != 0 {

					if lastListenedSong.Track != currentSong.Track || currentSong.IsRepeat {
						logger.Infof("Scrobbling new track for user. "+
							"Song change detected from %s to %s", lastListenedSong.Track, currentSong.Track)
						go lastfm.Scrobble(ctx, lastListenedSong, userId)
						go lastfm.UpdateNowPlaying(ctx, *currentSong, userId)
					}
				} else {
					logger.Warn(resp.Error)
				}

			}

		}
		resp := ctx.Database.Model(
			&types.CurrentListeningSongLocal{}).
			Where("id = ?", userId).
			Updates(map[string]interface{}{
				"track":  currentSong.Track,
				"artist": currentSong.Artist,
				"source": currentSong.Source,
				"url":    currentSong.Url,
			})
		err = resp.Error
		if err != nil || resp.RowsAffected == 0 {

			// always handle error like this, cause errors maybe happened when connection failed or something.
			// record not found...
			if resp.RowsAffected == 0 || gorm.IsRecordNotFoundError(err) {
				logger.Info("Creating a new record for current song")
				ctx.Database.Create(&types.CurrentListeningSongLocal{
					Id:       int(userId),
					Username: username,
					Track:    currentSong.Track,
					Artist:   currentSong.Artist,
					Source:   currentSong.Source,
					Url:      currentSong.Url,
				}) // create new record from newUser
			} else if err != nil {
				logger.Fatal(err)
			}

		}
		return c.SendStatus(fiber.StatusAccepted)
	})

	app.Get("/user/player/local/current_song", func(c *fiber.Ctx) error {
		user := c.Locals("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		userId := claims["id"].(float64)
		userInDatabase := types.CurrentListeningSongLocal{}
		resp := ctx.Database.First(&userInDatabase, "id = ?", userId)
		if resp.Error != nil {
			logger.Warn(resp.Error)
			return c.JSON(types.CurrentListeningSongLocal{})
		}
		return c.JSON(userInDatabase)
	})

	app.Get("/user/player/local/current_song/similar", func(c *fiber.Ctx) error {
		user := c.Locals("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		userId := claims["id"].(float64)
		userInDatabase := types.CurrentListeningSongLocal{}
		resp := ctx.Database.First(&userInDatabase, "id = ?", userId)
		if resp.Error != nil || resp.RowsAffected == 0 || userInDatabase.Track == "" {
			logger.Warn(resp.Error)
			return c.SendStatus(fiber.StatusNotFound)
		}

		similar, err := ctx.LastFm.Track.GetSimilar(map[string]interface{}{
			"artist": userInDatabase.GetCleanedArtistName(),
			"track":  userInDatabase.Track,
		})
		if err != nil {
			return err
		}

		songs := make([]types.SongMeta, len(similar.Tracks))

		for i := range similar.Tracks {
			trackMeta := similar.Tracks[i]
			track := trackMeta.Name
			artist := trackMeta.Artist.Name
			albumArt := ""
			if len(trackMeta.Images) != 0 {
				albumArt = trackMeta.Images[len(trackMeta.Images)-1].Url
			}
			songs = append(songs,
				types.SongMeta{
					Track:      track,
					Artist:     artist,
					Source:     "last.fm",
					Url:        trackMeta.Url,
					AlbumArt:   albumArt,
					Mbid:       trackMeta.Mbid,
					ArtistMbid: trackMeta.Artist.Mbid,
				})
		}

		return c.JSON(songs)
	})

	app.Get("/user/player/local/current_song/love", func(c *fiber.Ctx) error {
		user := c.Locals("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		userId := claims["id"].(float64)
		userInDatabase := types.CurrentListeningSongLocal{}
		resp := ctx.Database.First(&userInDatabase, "id = ?", userId)
		if resp.Error != nil || resp.RowsAffected == 0 {
			// user is not listening to any songs
			return c.SendStatus(fiber.StatusNotFound)
		}
		resp = ctx.Database.First(
			&types.LikedSong{}, "id = ? AND track = ? AND artist = ?",
			userId, userInDatabase.Track, userInDatabase.Artist)

		if resp.Error != nil || resp.RowsAffected == 0 {
			// this is the first attempt to like this
			likedTrack := types.LikedSong{Track: userInDatabase.Track, Artist: userInDatabase.Artist, Id: int(userId)}
			resp = ctx.Database.Create(likedTrack)
			if resp.Error != nil {
				logger.Warn(resp.Error)
				return c.SendStatus(fiber.StatusInternalServerError)
			} else {
				return c.SendStatus(fiber.StatusAccepted)
			}
		} else {
			return c.SendStatus(fiber.StatusAlreadyReported)
		}
	})

	app.Get("/user/player/local/current_song/unlove", func(c *fiber.Ctx) error {
		user := c.Locals("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		userId := claims["id"].(float64)
		userInDatabase := types.CurrentListeningSongLocal{}
		resp := ctx.Database.First(&userInDatabase, "id = ?", userId)
		if resp.Error != nil || resp.RowsAffected == 0 {
			// user is not listening to any songs
			return c.SendStatus(fiber.StatusNotFound)
		}

		resp = ctx.Database.First(
			&types.LikedSong{}, "id = ? AND track = ? AND artist = ?",
			userId, userInDatabase.Track, userInDatabase.Artist)
		return nil
	})

	app.Get("/user/player/local/current_song/lyrics", func(c *fiber.Ctx) error {
		user := c.Locals("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		userId := claims["id"].(float64)
		userInDatabase := types.CurrentListeningSongLocal{}
		resp := ctx.Database.First(&userInDatabase, "id = ?", userId)
		if resp.Error != nil || resp.RowsAffected == 0 {
			// user is not listening to any songs
			return c.SendStatus(fiber.StatusNotFound)
		}

		lyrics, err := sl.GetLyrics(slTypes.Song{
			Track:  userInDatabase.Track,
			Artist: userInDatabase.Artist,
		})
		logger.Infof("Request for lyrics: %s", lyrics)
		if err != nil {
			return err
		}
		return c.SendString(lyrics)

	})

	app.Get("/user/dot/all", func(c *fiber.Ctx) error {
		user := c.Locals("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		username := claims["user"].(string)
		var friends []types.Dot
		resp := ctx.Database.First(&friends, "username = ?", username)
		if resp.Error != nil {
			logger.Warn(resp.Error)
			return c.JSON([]types.Dot{})
		}
		return c.JSON(friends)
	})

	app.Post("/user/dot/add", func(c *fiber.Ctx) error {
		// data:Dot
		user := c.Locals("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		username := claims["user"].(string)

		var dot types.Dot
		err := json.Unmarshal(c.Body(), &dot)
		if err != nil {
			return err
		}
		dotInDatabase := types.Dot{}

		if dot.DotUsername == "" {
			return c.SendStatus(fiber.StatusForbidden)
		}

		// check if the user is already a dot
		resp := ctx.Database.First(&dotInDatabase, "username = ? AND dot_username = ?", username, dot.DotUsername)
		if resp.Error == nil {
			// the user already exists hmmmm
			logger.Warnf("%s is already connected with %s", username, dot.DotUsername)
			return c.SendStatus(fiber.StatusAlreadyReported)
		}

		parsedDot, err := helpers.ParseFullUserId(dot.DotUsername)
		if err != nil {
			// the user id is invalid
			logger.Warnf("%s attempted to make connect with dot %s which is an invalid id ")
			return err
		}
		if parsedDot.Homeserver == ctx.Config.Server.Name {
			userExists, err := isUserExists(parsedDot.Username)
			if err != nil {
				return err
			}
			if !userExists {
				return c.SendStatus(fiber.StatusBadRequest)
			}
		} else {
			return c.SendStatus(fiber.StatusNotImplemented)
		}

		count := -1
		ctx.Database.Model(&types.Dot{}).Count(&count)
		if count == -1 {
			err := errors.New("failed to get count of the number of records in friends table")
			logger.Warn(err)
			return err
		}
		resp = ctx.Database.Create(&types.Dot{Username: username, Id: count, DotUsername: dot.DotUsername})
		err = resp.Error
		if err != nil {
			return err
		}
		return c.SendStatus(fiber.StatusAccepted)
	})

	app.Post("/user/dot/remove", func(c *fiber.Ctx) error {
		user := c.Locals("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		username := claims["user"].(string)

		var dot types.Dot
		err := json.Unmarshal(c.Body(), &dot)
		if err != nil {
			return err
		}
		friendInDatabase := types.Dot{}

		// check if the user is already a friend
		resp := ctx.Database.First(&friendInDatabase, "username = ? AND friend_username = ?", username, dot)
		if resp.Error != nil {
			// the user already exists hmmmm
			logger.Warnf("%s is not connected with %s", username, dot.DotUsername)
			return errors.New("not connected yet")
		}

		// ctx.Database.Delete()
		return c.SendStatus(fiber.StatusInternalServerError)

		/*err = resp.Error
		if err != nil || resp.RowsAffected == 0 {

			// always handle error like this, cause errors maybe happened when connection failed or something.
			// record not found...
			if resp.RowsAffected == 0 || gorm.IsRecordNotFoundError(err) {
				logger.Info("Creating a new record for current song")
				ctx.Database.Create(&types.CurrentListeningSongLocal{
					Id:       int(userId),
					Username: username,
					Track:    currentSong.Track,
					Artist:   currentSong.Artist}) // create new record from newUser
			} else if err != nil {
				logger.Fatal(err)
			}

		}*/
	})

	app.Get("/dot/current_song", func(c *fiber.Ctx) error {
		user := c.Locals("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		username := claims["user"].(string)

		var dot types.Dot
		err := json.Unmarshal(c.Body(), &dot)
		if err != nil {
			return err
		}
		dotInDatabase := types.Dot{}

		// check if the user is already a dot
		resp := ctx.Database.First(&dotInDatabase, "username = ? AND dot_username = ?", username, dot.DotUsername)
		if resp.Error != nil {
			// the user is not a friend
			return c.SendStatus(fiber.StatusForbidden)
		}

		targetUser, err := helpers.ParseFullUserId(dot.DotUsername)
		logger.Info(targetUser)
		logger.Info(dot.DotUsername)
		if err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		// the homeserver of the user matches the homeserver of the friend
		// so we need to retrieve the data from the database and send it off
		if targetUser.Homeserver == ctx.Config.Server.Name {
			targetUserInDatabase := types.CurrentListeningSongLocal{}
			resp := ctx.Database.First(&targetUserInDatabase, "username = ?", targetUser.Username)
			if resp.Error != nil {
				logger.Warn(resp.Error)
				return c.JSON(types.CurrentListeningSongLocal{})
			}
			return c.JSON(targetUserInDatabase)
		}

		// federation has not been implemented yet
		// TODO: do this later
		return c.SendStatus(fiber.StatusNotImplemented)
	})

	app.Get("/connect/spotify", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNotImplemented)
	})

	if ctx.LastFm == nil {
		return app, nil
	}

	app.Get("/connect/lastfm", func(c *fiber.Ctx) error {
		token, err := ctx.LastFm.GetToken()
		if err != nil {
			return err
		}

		user := c.Locals("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		userId := claims["id"].(float64)
		username := claims["user"].(string)

		go lastfm.StoreAuthToken(ctx, username, userId, token)
		return c.JSON(map[string]string{
			"redirect": ctx.LastFm.GetAuthTokenUrl(token),
		})

	})

	return app, nil
}

func loginJWT(c *fiber.Ctx, user types.UserAccount, signingKey string) error {

	// Create token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = user.Id
	claims["user"] = user.Username
	claims["tgid"] = user.TelegramId
	claims["exp"] = time.Now().Add(time.Hour * 14 * 24).Unix()

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte(signingKey))
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(fiber.Map{"token": t})
}

func accessible(c *fiber.Ctx) error {
	return c.SendString("Hello world. I am alive. 👀")
}
