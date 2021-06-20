package types


// Dot is referred to as a individual on the Lyrix Fediverse
type Dot struct {
	Id             int    `gorm:"primary_key"`
	Username       string `json:"username,omitempty"`
	DotUsername string `json:"dot_username"`
}
