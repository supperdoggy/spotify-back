package structs

import globalStructs "github.com/supperdoggy/spotify-web-project/spotify-globalStructs"

type CreateNewSongReq struct {
	SongData []byte `json:"song_data"`
	globalStructs.Song
}
