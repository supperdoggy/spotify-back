package handlers

import (
	"encoding/json"
	structs2 "github.com/supperdoggy/spotify-web-project/spotify-auth/shared/structs"
	"github.com/supperdoggy/spotify-web-project/spotify-back/internal/service"
	"github.com/supperdoggy/spotify-web-project/spotify-back/internal/utils"
	"github.com/supperdoggy/spotify-web-project/spotify-back/shared/structs"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
)

type Handlers struct {
	logger *zap.Logger
	s      service.IService
}

func NewHandlers(l *zap.Logger, s service.IService) *Handlers {
	return &Handlers{logger: l, s: s}
}

func (h *Handlers) InitHandlers() {
	http.HandleFunc("/", h.GetSegment)
	http.HandleFunc("/api/v1/newsong", h.createNewSong)
	http.HandleFunc("/allsongs", h.getSongs)
	http.HandleFunc("/login", h.Login)
	http.HandleFunc("/register", h.Register)
}

// addHeaders will act as middleware to give us CORS support
func addHeaders(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		h.ServeHTTP(w, r)
	}
}

func (h *Handlers) GetSegment(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Access-Control-Allow-Origin", "*")
	// todo add check for auth access !!!
	id := request.RequestURI[1:]
	resp, err := h.s.GetSegment(id)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Write(resp)
}

func (h *Handlers) getSongs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var status = http.StatusOK
	// getNames
	resp, err := h.s.GetAllSongs()
	if err != nil {
		h.logger.Error("error getting all songs", zap.Error(err))
		status = http.StatusBadRequest
	}
	data, err := json.Marshal(resp)
	if err != nil {
		h.logger.Error("error marshaling response", zap.Error(err), zap.Any("data", resp))
		return
	}

	w.WriteHeader(status)
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

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	var req structs2.LoginReq
	var resp structs2.LoginResp
	err := utils.ParseJson(r, &req)
	if err != nil {
		h.logger.Error("error reading body", zap.Error(err))
		resp.Error = err.Error()
		utils.SendJson(w, resp, http.StatusBadRequest)
		return
	}

	resp, err = h.s.Login(req)
	if err != nil {
		h.logger.Error("gor Login() error", zap.Error(err))
		utils.SendJson(w, resp, http.StatusBadRequest)
		return
	}
	utils.SendJson(w, resp, http.StatusOK)
}

func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	var req structs2.RegisterReq
	var resp structs2.NewTokenResp
	err := utils.ParseJson(r, &req)
	if err != nil {
		h.logger.Error("error reading body", zap.Error(err))
		resp.Error = err.Error()
		utils.SendJson(w, resp, http.StatusBadRequest)
		return
	}

	resp, err = h.s.Register(req)
	if err != nil {
		h.logger.Error("gor Register() error", zap.Error(err))
		utils.SendJson(w, resp, http.StatusBadRequest)
		return
	}
	utils.SendJson(w, resp, http.StatusOK)
}
