package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	globalStructs "github.com/supperdoggy/spotify-web-project/spotify-globalStructs"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
)

func CreateMP3File(name string, data []byte) error {
	outputfile, err := os.Create("example/mp3/" + name + ".mp3")
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(data)
	_, err = io.Copy(outputfile, buf)
	return err
}

func ConvMp3ToM3U8(logger *zap.Logger, filename, m3p8 string) (m3u8Data *globalStructs.SongData, tsData []globalStructs.SongData, err error) {
	args := []string{"example/create.sh", "example/mp3/" + filename, "example/songs/" + m3p8 + ".m3u8", "example/songs/" + m3p8 + "_%03d.ts"}
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
		ID:   m3p8 + ".m3u8",
		Data: data,
	}

	//out = string(b)
	tsData = []globalStructs.SongData{}
	i := 0
	for {
		path := fmt.Sprintf("example/songs/%s_%03d.ts", m3p8, i)
		data, err := os.ReadFile(path)
		if err != nil {
			break
		}
		songData := globalStructs.SongData{
			ID:   fmt.Sprintf("%s_%03d.ts", m3p8, i),
			Data: data,
		}

		tsData = append(tsData, songData)
		i++

		err = os.Remove(path)
		if err != nil {
			break
		}
	}

	err = os.Remove(fmt.Sprintf("example/songs/%s.m3u8", m3p8))
	if err != nil {
		return nil, nil, err
	}

	if len(tsData) == 0 {
		return nil, nil, errors.New("tsData len is 0")
	}

	return
}

func SendRequest(req interface{}, method, url string, resp interface{}) error {
	var respFromServer *http.Response
	var data []byte
	var err error

	if req == nil {
		data, err = json.Marshal(req)
		if err != nil {
			return err
		}
	}
	switch method {
	case "post":
		respFromServer, err = http.Post(url, "application/json", bytes.NewBuffer(data))
	case "get":
		respFromServer, err = http.Get(url)
	}
	if err != nil {
		return err
	}
	defer respFromServer.Body.Close()

	data, err = ioutil.ReadAll(respFromServer.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return err
	}
	return nil
}

func SendJson(w http.ResponseWriter, obj interface{}, code int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

func ParseJson(r *http.Request, obj interface{}) error {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, obj)
}
