package utils

import (
	"errors"
	"fmt"
	globalStructs "github.com/supperdoggy/spotify-web-project/spotify-globalStructs"
	"go.uber.org/zap"
	"os"
	"os/exec"
)

func ConvMp3ToM3U8(logger *zap.Logger, filename, m3p8 string) (m3u8Data *globalStructs.SongData, tsData []globalStructs.SongData, err error) {
	args := []string{"example/create.sh", filename, "example/songs/"+m3p8+".m3u8", "example/songs/"+m3p8+"_%03d.ts"}
	cmd := exec.Command("/bin/sh", args...)
	_, err = cmd.CombinedOutput()
	if err != nil {
		logger.Error("error getting output", zap.Error(err))
		return nil, nil, err
	}

	data, err := os.ReadFile(fmt.Sprintf("example/songs/%s.m3u8", m3p8))
	if err != nil {
		logger.Error("error reading m3u8 file", zap.Error(err))
		return nil, nil, err
	}
	m3u8Data = &globalStructs.SongData{
		ID: m3p8+".m3u8",
		Data: data,
	}

	//out = string(b)
	tsData = []globalStructs.SongData{}
	i := 0
	for {
		data, err := os.ReadFile(fmt.Sprintf("example/songs/%s_%03d.ts", m3p8, i))
		if err != nil {
			break
		}
		songData := globalStructs.SongData{
			ID: fmt.Sprintf("%s_%03d.ts", m3p8, i),
			Data: data,
		}
		tsData = append(tsData, songData)
		i++
	}

	if len(tsData) == 0 {
		return nil, nil, errors.New("tsData len is 0")
	}

	return
}
