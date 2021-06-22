package types

import (

	sl "github.com/srevinsaju/swaglyrics-go"
	slTypes "github.com/srevinsaju/swaglyrics-go/types"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"time"
)

type User struct {
	Username string `json:"username"`
	Homeserver string `json:"homeserver"`
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


// CurrentListeningSongLocal represents a song which the user
// is currently listening on the local player
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


type LikedSong struct {
	Id int `gorm:"primary_key" json:"id"`
	Track string `json:"track"`
	Artist string `json:"artist"`
	CreatedAt time.Time
	UpdatedAt time.Time
}


func (sm CurrentListeningSongLocal) GetFirstArtist() string {
	if strings.Contains(sm.Artist, ", ") {
		firstArtist := strings.Split(sm.Artist, ",")[0]
		return strings.Trim(firstArtist, " ")
	}
	return sm.Artist
}

func (sm CurrentListeningSongLocal) GetCleanedArtistName() string {
	artist := strings.Replace(sm.GetFirstArtist(), " - Topic", "", -1)
	artist = strings.Replace(artist, " - Music", "", -1)
	strippedArtist := sl.Stripper(slTypes.Song{Artist: sm.Artist, Track: ""})
	strippedArtist = strings.TrimRight(strippedArtist, "-")
	if strippedArtist == "" {
		return artist
	}
	return strippedArtist
}
