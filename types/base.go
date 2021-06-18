package types

import (
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type SongMeta struct {
	Track  string `json:"track"`
	Artist string `json:"artist"`
	Source string `json:"source,omitempty"`
	Url    string `json:"url,omitempty"`
}

type UserAccount struct {
	Id             int    `gorm:"primary_key" json:"id"`
	Username       string `json:"username"`
	TelegramId     int    `json:"-"`
	HashedPassword string `json:"-"`
}

type UserAccountRegister struct {
	Id         int    `json:"id"`
	Username   string `json:"username"`
	TelegramId int    `json:"telegram_id"`
	Password   string `json:"password"`
}

type UserAccountLogin struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (uar UserAccountRegister) Hash() (*UserAccount, error) {
	rawHash, err := bcrypt.GenerateFromPassword([]byte(uar.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return &UserAccount{Id: uar.Id, Username: uar.Username, TelegramId: uar.TelegramId, HashedPassword: string(rawHash)}, nil
}

type CurrentListeningSongLocal struct {
	Id       int    `gorm:"primary_key" json:"id"`
	Username string `json:"username"`

	Track  string `json:"track"`
	Artist string `json:"artist"`

	Source string `json:"source,omitempty"`
	Url    string `json:"url,omitempty"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

type SpotifyAuthToken struct {
	Id       int    `gorm:"primary_key" json:"id"`
	Username string `json:"username"`
	Token    string `json:"token"`
}

type SpotifyAuthTokenRegisterRequest struct {
	Token string `json:"spotify_token"`
}

type Context struct {
	Database *gorm.DB
}

type Friend struct {
	Id             int    `gorm:"primary_key"`
	Username       string `json:"username,omitempty"`
	FriendUsername string `json:"friend_username"`
}
