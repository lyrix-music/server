package types

// LastFmAuthToken stores the latest token in the backend database
type LastFmAuthToken struct {
	Id       int    `gorm:"primary_key" json:"id" yaml:"id"`
	Username string `json:"username" yaml:"username"`
	Token    string `json:"token" yaml:"token"`
}

// LastFmSessionKey stores the latest token in the backend database
type LastFmSessionKey struct {
	Id         int    `gorm:"primary_key" json:"id" yaml:"id"`
	Username   string `json:"username" yaml:"username"`
	SessionKey string `json:"session_key" yaml:"session_key"`
}

// LastFmAuthTokenRegisterRequest is used to receive the token from
// the web interface.
type LastFmAuthTokenRegisterRequest struct {
	Token string `json:"lastfm_token"`
}
