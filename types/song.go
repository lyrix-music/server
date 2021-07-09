package types

import "strings"

// SongMeta represents a song on the Lyrix Database
type SongMeta struct {
	Track  string `json:"track"`
	Artist string `json:"artist"`
	Source string `json:"source,omitempty"`
	Url    string `json:"url,omitempty"`
	Scrobble bool `json:"scrobble,omitempty"`
	AlbumArt string `json:"album_art"`
	Mbid string `json:"mbid,omitempty"`
	ArtistMbid string `json:"artist_mbid,omitempty"`
	IsRepeat bool `json:"is_repeat"`
}

func (sm SongMeta) GetFirstArtist() string {
	if strings.Contains(sm.Artist, ", ") {
		firstArtist := strings.Split(sm.Artist, ",")[0]
		return strings.Trim(firstArtist, " ")
	}
	return sm.Artist
}

func (sm SongMeta) GetCurrentListeningSong() CurrentListeningSongLocal {
	return CurrentListeningSongLocal{Track: sm.Track, Artist: sm.Artist}
}
