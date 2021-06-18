package routes

import (
	"encoding/json"
	"errors"
	"os"

	"time"

	"github.com/jinzhu/gorm"
	"github.com/srevinsaju/lyrix/backend/config"
	"github.com/srevinsaju/lyrix/backend/types"
	"github.com/withmandala/go-log"
	"golang.org/x/crypto/bcrypt"

	jwt "github.com/form3tech-oss/jwt-go"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	mwLogger "github.com/gofiber/fiber/v2/middleware/logger"
	jwtware "github.com/gofiber/jwt/v2"
)

var logger = log.New(os.Stdout)

func Initialize(cfg config.Config, ctx *types.Context) (*fiber.App, error) {

	app := fiber.New(fiber.Config{DisableStartupMessage: true})

	logger.Infof("Preparing listener")

	app.Use(mwLogger.New())
	// Register
	app.Post("/register", func(c *fiber.Ctx) error {
		user := &types.UserAccountRegister{}

		err := json.Unmarshal(c.Body(), user)
		if err != nil {
			return err
		}
		if user.Username == "" {
			return c.SendStatus(fiber.StatusUnauthorized)
		}
		logger.Info(string(c.Body()), user)

		userInDatabase := types.UserAccount{}
		ctx.Database.Where("username = ?", user.Username).Find(&userInDatabase)
		if userInDatabase.Username == user.Username {
			return c.SendStatus(fiber.StatusUnauthorized)
		}
		count := -1
		ctx.Database.Model(&types.UserAccount{}).Count(&count)

		logger.Info(userInDatabase)

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

	app.Use("/login", limiter.New())
	// Login route
	app.Post("/login", func(c *fiber.Ctx) error {
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

	// Unauthenticated route
	app.Get("/", accessible)

	// JWT Middleware
	app.Use("/user", "/people", jwtware.New(jwtware.Config{
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

	app.Post("/user/player/spotify/token", func(c *fiber.Ctx) error {
		spotifyToken := &types.SpotifyAuthTokenRegisterRequest{}

		err := json.Unmarshal(c.Body(), spotifyToken)
		if err != nil {
			return err
		}

		user := c.Locals("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		userId := claims["id"].(float64)
		username := claims["user"].(string)
		if err := ctx.Database.Model(&types.SpotifyAuthToken{}).Where("username = ?", username).Update("token", spotifyToken.Token).Error; err != nil {
			// always handle error like this, cause errors maybe happened when connection failed or something.
			// record not found...
			if gorm.IsRecordNotFoundError(err) {
				ctx.Database.Create(&types.SpotifyAuthToken{Id: int(userId), Username: username, Token: spotifyToken.Token}) // create new record from newUser
			}
		}

		return c.SendStatus(fiber.StatusAccepted)
	})

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
		currentSong := &types.SongMeta{}

		err := json.Unmarshal(c.Body(), currentSong)
		if err != nil {
			return err
		}

		user, ok := c.Locals("user").(*jwt.Token)
		if !ok {
			return fiber.StatusUnauthorized
		}
		claims := user.Claims.(jwt.MapClaims)
		userId := claims["id"].(float64)
		username := claims["user"].(string)
		logger.Info("Attempting to update")
		// logger.Info(currentSong)

		if currentSong.Artist == "" || currentSong.Track == "" {
			err := ctx.Database.Model(
				&types.CurrentListeningSongLocal{}).
				Where("id = ?", userId).
				Updates(map[string]interface{}{"track": "", "artist": ""}).
				Error
			if err != nil {
				logger.Warn(err)
				return c.SendStatus(fiber.StatusNotFound)
			}
			return c.SendStatus(fiber.StatusAccepted)
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
					Artist:   currentSong.Artist}) // create new record from newUser
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

	app.Get("/user/friends", func(c *fiber.Ctx) error {
		user := c.Locals("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		username := claims["user"].(string)
		var friends []types.Friend
		resp := ctx.Database.First(&friends, "username = ?", username)
		if resp.Error != nil {
			logger.Warn(resp.Error)
			return c.JSON([]types.Friend{})
		}
		return c.JSON(friends)
	})

	app.Post("/user/friend/add", func(c *fiber.Ctx) error {
		user := c.Locals("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		username := claims["user"].(string)

		var friend types.Friend
		err := json.Unmarshal(c.Body(), &friend)
		if err != nil {
			return err
		}
		friendInDatabase := types.Friend{}

		// check if the user is already a friend
		resp := ctx.Database.First(&friendInDatabase, "username = ? AND friend_username = ?", username, friend)
		if resp.Error == nil {
			// the user already exists hmmmm
			logger.Warnf("%s is already connected with %s", username, friend.FriendUsername)
			return c.SendStatus(fiber.StatusAlreadyReported)
		}

		count := -1
		ctx.Database.Model(&types.Friend{}).Count(&count)
		if count == -1 {
			err := errors.New("failed to get count of the number of records in friends table")
			logger.Warn(err)
			return err
		}
		resp = ctx.Database.Create(&types.Friend{Username: username, Id: count, FriendUsername: friend.FriendUsername})
		err = resp.Error
		if err != nil {
			return err
		}
		return c.SendStatus(fiber.StatusAccepted)
	})

	app.Post("/user/friend/remove", func(c *fiber.Ctx) error {
		user := c.Locals("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		username := claims["user"].(string)

		var friend types.Friend
		err := json.Unmarshal(c.Body(), &friend)
		if err != nil {
			return err
		}
		friendInDatabase := types.Friend{}

		// check if the user is already a friend
		resp := ctx.Database.First(&friendInDatabase, "username = ? AND friend_username = ?", username, friend)
		if resp.Error != nil {
			// the user already exists hmmmm
			logger.Warnf("%s is not connected with %s", username, friend.FriendUsername)
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

	app.Get("/people/all", func(c *fiber.Ctx) error {
		user := c.Locals("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		userId := claims["id"].(float64)
		var usersInDatabase []types.UserAccount
		resp := ctx.Database.Where("id != ?", userId).Select("username", "id").Find(&usersInDatabase)
		if resp.Error != nil {
			return c.JSON([]types.UserAccount{})
		}
		return c.JSON(usersInDatabase)
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
