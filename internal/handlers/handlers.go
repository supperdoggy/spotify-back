package handlers

import (
	"encoding/json"
	globalStructs "github.com/supperdoggy/spotify-web-project/spotify-globalStructs"
	"go.uber.org/zap"
	"net/http"
)

type Handlers struct {
	logger *zap.Logger
}

func NewHandlers(l *zap.Logger) *Handlers {
	return &Handlers{logger: l}
}

func (h *Handlers) InitHandlers() {
	const songsDir = "example"
	http.Handle("/", addHeaders(http.FileServer(http.Dir(songsDir))))
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
			Path: "localhost:8080/ex.m3u8",
		},
		{
			Name: "шото там на девятке",
			Band: "подруга гспд",
			Album: "какой-то дебютный",
			Path: "localhost:8080/ex1.m3u8",
		},
	}
	json.NewEncoder(w).Encode(resp)
}
