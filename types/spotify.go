package types


// SpotifyAuthToken stores the latest token in the backend database
type SpotifyAuthToken struct {
	Id       int    `gorm:"primary_key" json:"id"`
	Username string `json:"username"`
	Token    string `json:"token"`
}


// SpotifyAuthTokenRegisterRequest is used to receive the token from
// the web interface.
type SpotifyAuthTokenRegisterRequest struct {
	Token string `json:"spotify_token"`
}