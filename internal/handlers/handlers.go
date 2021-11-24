package handlers

import (
	"encoding/json"
	"github.com/supperdoggy/spotify-web-project/spotify-back/internal/service"
	"github.com/supperdoggy/spotify-web-project/spotify-back/internal/structs"
	"github.com/supperdoggy/spotify-web-project/spotify-back/internal/utils"
	globalStructs "github.com/supperdoggy/spotify-web-project/spotify-globalStructs"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
)

type Handlers struct {
	logger *zap.Logger
	s service.IService
}

func NewHandlers(l *zap.Logger, s service.IService) *Handlers {
	return &Handlers{logger: l, s: s}
}

func (h *Handlers) InitHandlers() {
	count := -1
	const songsDir = "example/songs"
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		m3u8Data, tsData, err := utils.ConvMp3ToM3U8(h.logger, "example/mp3/ex.mp3", "ex")
		if err != nil {
			panic(err.Error())
		}
		writer.Header().Set("Access-Control-Allow-Origin", "*")
		if count == -1 {
			writer.Write(m3u8Data.Data)
			count++
			return
		}
		//TODO create endpoint which converts all mp3 int m3u8 and saves it into db then find why do we have a problem when id does not call for ex_002 ex_004 and ex_005
		writer.Write(tsData[count].Data)
	})
	http.HandleFunc("/api/v1/newsong", h.createNewSong)
	http.HandleFunc("/allsongs", h.getSongs)
}

// addHeaders will act as middleware to give us CORS support
func addHeaders(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		h.ServeHTTP(w, r)
	}
}

func (h *Handlers) getSongs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// getNames
	resp := []globalStructs.Song{
		{
			Name: "lipsi ha",
			Band: "Instasamka",
			Album: "Money day",
			Path: "http://localhost:8080/ex.m3u8",
		},
		{
			Name: "шото там на девятке",
			Band: "подруга гспд",
			Album: "какой-то дебютный",
			Path: "http://localhost:8080/ex1.m3u8",
		},
	}
	data, err := json.Marshal(resp)
	if err != nil {
		h.logger.Error("error marshaling response", zap.Error(err), zap.Any("data", resp))
		return
	}
	w.Write(data)
}

func (h *Handlers) createNewSong(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	var req structs.CreateNewSongReq
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("error reading body", zap.Error(err))
		w.Write([]byte("{'error':'error parsing body'}"))
		return
	}

	err = json.Unmarshal(data, &req)
	if err != nil {
		h.logger.Error("error reading body", zap.Error(err))
		w.Write([]byte("{'error':'error unmarshalling req'}"))
		return
	}

	err = h.s.CreateNewSong(req)

}
