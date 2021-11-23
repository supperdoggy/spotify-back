package service

import (
	"errors"
	"github.com/supperdoggy/spotify-web-project/spotify-back/internal/structs"
	"go.uber.org/zap"
)

type IService interface {
	CreateNewSong(req structs.CreateNewSongReq) error
}

type Service struct {
	logger *zap.Logger
}

func NewService(l *zap.Logger) IService {
	return &Service{logger: l}
}

func (s *Service) CreateNewSong(req structs.CreateNewSongReq) error {
	if req.SongData == nil || len(req.SongData) == 0 {
		return errors.New("no song data provided")
	}



	return nil
}
