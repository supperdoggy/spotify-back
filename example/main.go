package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func main() {
	args := []string{"./create.sh", "ex.mp3", "ex.m3u8", "ex_%03d.ts"}
	output, err := RunCMD("/bin/sh", args, true)
	if err != nil {
		fmt.Println("Error:", output)
	} else {
		fmt.Println("Result:", output)
	}

}

// RunCMD is a simple wrapper around terminal commands
func RunCMD(path string, arg []string, debug bool) (out string, err error) {

	cmd := exec.Command(path, arg...)

	var b []byte
	b, err = cmd.CombinedOutput()
	out = string(b)

	if debug {
		fmt.Println(strings.Join(cmd.Args[:], " "))

		if err != nil {
			fmt.Println("RunCMD ERROR")
			fmt.Println(out)
		}
	}

	return
}
