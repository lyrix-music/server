package types

import "strings"

// SongMeta represents a song on the Lyrix Database
type SongMeta struct {
	Track  string `json:"track"`
	Artist string `json:"artist"`
	Source string `json:"source,omitempty"`
	Url    string `json:"url,omitempty"`
}

func (sm SongMeta) GetFirstArtist() string {
	if strings.Contains(sm.Artist, ", ") {
		firstArtist := strings.Split(sm.Artist, ",")[0]
		return strings.Trim(firstArtist, " ")
	}
	return sm.Artist
}


