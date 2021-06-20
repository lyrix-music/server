package types


// LastFmAuthToken stores the latest token in the backend database
type LastFmAuthToken struct {
	Id       int    `gorm:"primary_key" json:"id"`
	Username string `json:"username"`
	Token    string `json:"token"`
}



// LastFmAuthTokenRegisterRequest is used to receive the token from
// the web interface.
type LastFmAuthTokenRegisterRequest struct {
	Token string `json:"lastfm_token"`
}