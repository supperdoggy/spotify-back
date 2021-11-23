package utils

import (
	"os/exec"
)

func ConvMp3ToM3U8(filename, m3p8 string) error {
	args := []string{"example/create.sh", filename, "example/songs/"+m3p8+".m3u8", "example/songs/"+m3p8+"_%03d.ts"}
	cmd := exec.Command("/bin/sh", args...)
	_, err := cmd.CombinedOutput()
	//out = string(b)

	return err
}
