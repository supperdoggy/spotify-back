package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/floyernick/fleep-go"
	structs2 "github.com/supperdoggy/spotify-web-project/spotify-auth/shared/structs"
	"github.com/supperdoggy/spotify-web-project/spotify-back/internal/utils"
	"github.com/supperdoggy/spotify-web-project/spotify-back/shared/structs"
	dbStructs "github.com/supperdoggy/spotify-web-project/spotify-db/shared/structs"
	structsDB "github.com/supperdoggy/spotify-web-project/spotify-db/shared/structs"
	globalStructs "github.com/supperdoggy/spotify-web-project/spotify-globalStructs"
	"github.com/u2takey/go-utils/rand"
	"go.uber.org/zap"
	"gopkg.in/night-codes/types.v1"
	"io/ioutil"
	"net/http"
	"time"
)

type IService interface {
	CreateNewSong(req structs.CreateNewSongReq) error
	GetAllSongs() (resp structsDB.GetAllSongsResp, err error)
	GetSegment(id string) ([]byte, error)
	Register(req structs2.RegisterReq) (resp structs2.NewTokenResp, err error)
	Login(req structs2.LoginReq) (resp structs2.LoginResp, err error)
	RemoveSongFromPlaylist(req structsDB.RemoveSongFromUserPlaylistReq) (resp structsDB.RemoveSongFromUserPlaylistResp, err error)
	GetUserPlaylists(req structsDB.GetUserAllPlaylistsReq) (resp structsDB.GetUserAllPlaylistsResp, err error)
	GetPlaylist(req structsDB.GetPlaylistReq) (resp structsDB.GetPlaylistResp, err error)
	NewPlaylist(req structsDB.NewPlaylistReq) (resp structsDB.NewPlaylistResp, err error)
	AddSongToPlaylist(req structsDB.AddSongToUserPlaylistReq) (resp structsDB.AddSongToUserPlaylistResp, err error)
}

type Service struct {
	logger *zap.Logger
}

func NewService(l *zap.Logger) IService {
	return &Service{logger: l}
}

func (s *Service) CreateNewSong(req structs.CreateNewSongReq) error {
	if req.SongData == nil || len(req.SongData) == 0 || req.Name == "" || req.Band == "" || req.Album == "" {
		return errors.New("fill all the fields")
	}

	info, err := fleep.GetInfo(req.SongData)
	if err != nil {
		return err
	}

	if !info.IsAudio() {
		return errors.New("file should be audio")
	}

	fileName := types.String(time.Now().UnixNano())
	err = utils.CreateMP3File(fileName, req.SongData)
	if err != nil {
		s.logger.Error("error creating new mp3 file", zap.Error(err))
		return err
	}

	m3h8, ts, err := utils.ConvMp3ToM3U8(s.logger, fileName+".mp3", fileName)
	if err != nil {
		s.logger.Error("error converting mp3 to m3u8", zap.Error(err))
		return err
	}

	song := globalStructs.Song{
		ID:          fileName,
		Name:        req.Name,
		Album:       req.Album,
		Band:        req.Band,
		ReleaseDate: req.ReleaseDate,
		Path:        fmt.Sprintf("http://localhost:8080/%s.m3u8", fileName),
	}

	var respFromDB dbStructs.AddSegmentsResp
	reqToDB := dbStructs.AddSegmentsReq{
		//UserID: userID
		Ts:       ts,
		M3H8:     *m3h8,
		SongData: song,
	}

	marshalled, err := json.Marshal(reqToDB)
	if err != nil {
		s.logger.Error("error marshaling req to db", zap.Error(err))
		return err
	}

	buf := bytes.NewBuffer(marshalled)

	resp, err := http.Post("http://localhost:8082/api/v1/addSegment", "application/json", buf)
	if err != nil {
		s.logger.Error("error making req to db", zap.Error(err))
		return err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		s.logger.Error("error reading body", zap.Error(err))
		return err
	}

	err = json.Unmarshal(data, &respFromDB)
	if err != nil {
		s.logger.Error("error unmarshaling answer", zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	if !respFromDB.OK {
		s.logger.Error("got error from db", zap.Any("error", respFromDB.Error))
		return errors.New(respFromDB.Error)
	}

	return nil
}

func (s *Service) GetAllSongs() (resp structsDB.GetAllSongsResp, err error) {
	err = utils.SendRequest(nil, "get", "http://localhost:8082/api/v1/allsongs", &resp)
	if err != nil {
		s.logger.Error("error sending request", zap.Error(err))
		resp.Error = err.Error()
		return
	}
	if resp.Error != "" {
		s.logger.Error("got error from db", zap.Any("error", resp.Error))
		return resp, errors.New(resp.Error)
	}

	return resp, err
}

func (s *Service) GetSegment(id string) ([]byte, error) {
	// CHECK TOKEN!!!!!
	req := structsDB.GetSegmentReq{
		ID: id,
	}

	marshalled, err := json.Marshal(req)
	if err != nil {
		s.logger.Error("cant marshall req", zap.Error(err))
		return nil, err
	}

	rawResult, err := http.Post("http://localhost:8082/api/v1/getsegment", "application/json", bytes.NewBuffer(marshalled))
	if err != nil {
		s.logger.Error("error making response to db", zap.Error(err), zap.Any("req", req))
		return nil, err
	}
	defer rawResult.Body.Close()

	data, err := ioutil.ReadAll(rawResult.Body)
	if err != nil {
		s.logger.Error("error reading result body", zap.Error(err), zap.Any("req", req))
		return nil, err
	}

	var resp structsDB.GetSegmentResp
	if err := json.Unmarshal(data, &resp); err != nil {
		s.logger.Error("error unmarshalling resp from db", zap.Error(err))
		return nil, err
	}

	if resp.Error != "" {
		s.logger.Error("got error from db", zap.Any("error", resp.Error))
		return nil, errors.New(resp.Error)
	}

	return resp.Segment.Data, nil
}


func (s *Service) Register(req structs2.RegisterReq) (resp structs2.NewTokenResp, err error) {
	if req.Password == "" || req.Email == "" {
		resp.Error = "fill all the fields"
		return resp, errors.New("fill all the fields")
	}

	var respFromAuth structs2.RegisterResp
	marshalled, err := json.Marshal(req)
	if err != nil {
		resp.Error = err.Error()
		return
	}

	respdata, err := http.Post("http://localhost:8083/api/v1/register", "application/json", bytes.NewBuffer(marshalled))
	if err != nil {
		resp.Error = err.Error()
		return
	}

	data, err := ioutil.ReadAll(respdata.Body)
	if err != nil {
		resp.Error = err.Error()
		return
	}

	err = json.Unmarshal(data, &respFromAuth)
	if err != nil {
		s.logger.Error("error sending request to auth", zap.Error(err))
		resp.Error = err.Error()
		return resp, errors.New("error making request")
	}

	if respFromAuth.Error != "" {
		s.logger.Error("got error from auth", zap.Any("error", respFromAuth.Error))
		resp.Error = respFromAuth.Error
		return resp, errors.New(respFromAuth.Error)
	}

	user := globalStructs.User{
		ID:         respFromAuth.UserID,
		Username:   rand.String(12),
		Email:      req.Email,
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		Statuses:   globalStructs.Statuses{},
		LastOnline: time.Now(),
	}
	var respFromDB structsDB.NewUserResp

	marshalled, err = json.Marshal(user)
	if err != nil {
		resp.Error = err.Error()
		return
	}

	respDB, err := http.Post("http://localhost:8082/api/v1/new_user", "application/json", bytes.NewBuffer(marshalled))
	if err != nil {
		resp.Error = err.Error()
		return
	}

	data, err = ioutil.ReadAll(respDB.Body)
	if err != nil {
		resp.Error = err.Error()
		return
	}

	err = json.Unmarshal(data, &respFromDB)
	if err != nil {
		resp.Error = err.Error()
		return
	}

	if !respFromDB.OK {
		s.logger.Error("got error from db", zap.Any("error", respFromDB.Error), zap.Any("user", user))
		resp.Error = respFromDB.Error
		return resp, err
	}

	resp.Token = respFromAuth.Token
	return resp, nil
}

func (s *Service) Login(req structs2.LoginReq) (resp structs2.LoginResp, err error) {
	if req.Email == "" || req.Password == "" {
		resp.Error = "fill all the fields"
		return resp, errors.New(resp.Error)
	}

	marshalled, err := json.Marshal(req)
	if err != nil {
		s.logger.Error("error marshalling data", zap.Error(err))
		resp.Error = err.Error()
		return
	}

	respdata, err := http.Post("http://localhost:8083/api/v1/login", "application/json", bytes.NewBuffer(marshalled))
	if err != nil {
		s.logger.Error("error making post request to auth", zap.Error(err))
		resp.Error = err.Error()
		return
	}

	data, err := ioutil.ReadAll(respdata.Body)
	if err != nil {
		s.logger.Error("error reading body", zap.Error(err))
		resp.Error = err.Error()
		return
	}

	err = json.Unmarshal(data, &resp)
	if err != nil {
		s.logger.Error("error unmarshalling data", zap.Error(err))
		resp.Error = err.Error()
		return
	}

	if resp.Error != "" {
		s.logger.Error("got error from auth", zap.Any("error", resp.Error))
		return resp, errors.New(resp.Error)
	}

	return resp, nil
}

func (s *Service) GetUserPlaylists(req structsDB.GetUserAllPlaylistsReq) (resp structsDB.GetUserAllPlaylistsResp, err error) {
	if req.UserID == "" {
		resp.Error = "you must fill all ids"
		return resp, errors.New(resp.Error)
	}

	data, err := json.Marshal(req)
	if err != nil {
		s.logger.Error("error marshalling requst",  zap.Error(err))
		resp.Error = err.Error()
		return resp, err
	}

	response, err := http.Post("http://localhost:8082/api/v1/user_playlists", "application/json", bytes.NewBuffer(data))
	if err != nil {
		s.logger.Error("error making post req to db", zap.Error(err))
		resp.Error = err.Error()
		return resp, err
	}

	data, err = ioutil.ReadAll(response.Body)
	if err != nil {
		s.logger.Error("error reading body", zap.Error(err))
		resp.Error = err.Error()
		return
	}

	err = json.Unmarshal(data, &resp)
	if err != nil {
		s.logger.Error("error unmarshalling resp", zap.Error(err), zap.Any("data", string(data)))
		resp.Error = err.Error()
		return
	}

	if resp.Error != "" {
		s.logger.Error("got error from db", zap.Any("error", resp.Error))
		return resp, errors.New(resp.Error)
	}

	return
}

func (s *Service) GetPlaylist(req structsDB.GetPlaylistReq) (resp structsDB.GetPlaylistResp, err error) {
	if req.PlaylistID == "" {
		resp.Error = "you must fill all ids"
		return resp, errors.New(resp.Error)
	}

	data, err := json.Marshal(req)
	if err != nil {
		s.logger.Error("error marshalling requst",  zap.Error(err))
		resp.Error = err.Error()
		return resp, err
	}

	response, err := http.Post("http://localhost:8082/api/v1/get_playlist ", "application/json", bytes.NewBuffer(data))
	if err != nil {
		s.logger.Error("error making post req to db", zap.Error(err))
		resp.Error = err.Error()
		return resp, err
	}

	data, err = ioutil.ReadAll(response.Body)
	if err != nil {
		s.logger.Error("error reading body", zap.Error(err))
		resp.Error = err.Error()
		return
	}

	err = json.Unmarshal(data, &resp)
	if err != nil {
		s.logger.Error("error unmarshalling resp", zap.Error(err), zap.Any("data", string(data)))
		resp.Error = err.Error()
		return
	}

	if resp.Error != "" {
		s.logger.Error("got error from db", zap.Any("error", resp.Error))
		return resp, errors.New(resp.Error)
	}

	return
}

func (s *Service) NewPlaylist(req structsDB.NewPlaylistReq) (resp structsDB.NewPlaylistResp, err error) {
	if req.PlaylistName == "" || req.UserID == "" {
		resp.Error = "you must fill all ids"
		return resp, errors.New(resp.Error)
	}

	data, err := json.Marshal(req)
	if err != nil {
		s.logger.Error("error marshalling requst",  zap.Error(err))
		resp.Error = err.Error()
		return resp, err
	}

	response, err := http.Post("http://localhost:8082/api/v1/new_playlist", "application/json", bytes.NewBuffer(data))
	if err != nil {
		s.logger.Error("error making post req to db", zap.Error(err))
		resp.Error = err.Error()
		return resp, err
	}

	data, err = ioutil.ReadAll(response.Body)
	if err != nil {
		s.logger.Error("error reading body", zap.Error(err))
		resp.Error = err.Error()
		return
	}

	err = json.Unmarshal(data, &resp)
	if err != nil {
		s.logger.Error("error unmarshalling resp", zap.Error(err), zap.Any("data", string(data)))
		resp.Error = err.Error()
		return
	}

	if resp.Error != "" {
		s.logger.Error("got error from db", zap.Any("error", resp.Error))
		return resp, errors.New(resp.Error)
	}

	return
}


func (s *Service) AddSongToPlaylist(req structsDB.AddSongToUserPlaylistReq) (resp structsDB.AddSongToUserPlaylistResp, err error) {
	if req.PlaylistID == "" || req.UserID == "" || req.SongID == "" {
		resp.Error = "you must fill all ids"
		return resp, errors.New(resp.Error)
	}

	data, err := json.Marshal(req)
	if err != nil {
		s.logger.Error("error marshalling requst",  zap.Error(err))
		resp.Error = err.Error()
		return resp, err
	}

	response, err := http.Post("http://localhost:8082/api/v1/add_song_playlist", "application/json", bytes.NewBuffer(data))
	if err != nil {
		s.logger.Error("error making post req to db", zap.Error(err))
		resp.Error = err.Error()
		return resp, err
	}

	data, err = ioutil.ReadAll(response.Body)
	if err != nil {
		s.logger.Error("error reading body", zap.Error(err))
		resp.Error = err.Error()
		return
	}

	err = json.Unmarshal(data, &resp)
	if err != nil {
		s.logger.Error("error unmarshalling resp", zap.Error(err), zap.Any("data", string(data)))
		resp.Error = err.Error()
		return
	}

	if resp.Error != "" {
		s.logger.Error("got error from db", zap.Any("error", resp.Error))
		return resp, errors.New(resp.Error)
	}

	return
}

func (s *Service) RemoveSongFromPlaylist(req structsDB.RemoveSongFromUserPlaylistReq) (resp structsDB.RemoveSongFromUserPlaylistResp, err error) {
	if req.PlaylistID == "" || req.UserID == "" || req.SongID == "" {
		resp.Error = "you must fill all ids"
		return resp, errors.New(resp.Error)
	}

	data, err := json.Marshal(req)
	if err != nil {
		s.logger.Error("error marshalling requst",  zap.Error(err))
		resp.Error = err.Error()
		return resp, err
	}

	response, err := http.Post("http://localhost:8082/api/v1/remove_song_playlist", "application/json", bytes.NewBuffer(data))
	if err != nil {
		s.logger.Error("error making post req to db", zap.Error(err))
		resp.Error = err.Error()
		return resp, err
	}

	data, err = ioutil.ReadAll(response.Body)
	if err != nil {
		s.logger.Error("error reading body", zap.Error(err))
		resp.Error = err.Error()
		return
	}

	err = json.Unmarshal(data, &resp)
	if err != nil {
		s.logger.Error("error unmarshalling resp", zap.Error(err), zap.Any("data", string(data)))
		resp.Error = err.Error()
		return
	}

	if resp.Error != "" {
		s.logger.Error("got error from db", zap.Any("error", resp.Error))
		return resp, errors.New(resp.Error)
	}

	return
}
