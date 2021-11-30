package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/floyernick/fleep-go"
	"github.com/supperdoggy/spotify-web-project/spotify-back/internal/utils"
	"github.com/supperdoggy/spotify-web-project/spotify-back/shared/structs"
	dbStructs "github.com/supperdoggy/spotify-web-project/spotify-db/shared/structs"
	structsDB "github.com/supperdoggy/spotify-web-project/spotify-db/shared/structs"
	globalStructs "github.com/supperdoggy/spotify-web-project/spotify-globalStructs"
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

	// todo save to db
	s.logger.Info("result", zap.Any("m3h8", m3h8), zap.Any("len ts", len(ts)))

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
	rawResp, err := http.Get("http://localhost:8082/api/v1/allsongs")
	if err != nil {
		s.logger.Error("error making request to db", zap.Error(err))
		resp.Error = err.Error()
		return resp, err
	}
	defer rawResp.Body.Close()

	data, err := ioutil.ReadAll(rawResp.Body)
	if err != nil {
		s.logger.Error("error reading body", zap.Error(err))
		resp.Error = err.Error()
		return resp, err
	}

	if err := json.Unmarshal(data, &resp); err != nil {
		s.logger.Error("error unmarshaling req", zap.Error(err))
		resp.Error = err.Error()
		return resp, err
	}

	if resp.Error != "" {
		s.logger.Error("got error from db", zap.Any("error", resp.Error))
		return resp, errors.New(resp.Error)
	}

	return resp, err
}

func (s *Service) GetSegment(id string) ([]byte, error) {
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
