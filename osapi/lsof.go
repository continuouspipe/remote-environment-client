package osapi

import (
	"os/exec"
	"strconv"
	"strings"
)

func GetCurrentlyOpenedFilesCount() (int, error) {
	lsof := exec.Command("lsof")
	wc := exec.Command("wc", "-l")
	outPipe, err := lsof.StdoutPipe()
	defer outPipe.Close()
	if err != nil {
		return 0, err
	}
	lsof.Start()
	wc.Stdin = outPipe
	out, err := wc.Output()
	if err != nil {
		return 0, err
	}
	outstr := string(out[:])
	count, err := strconv.Atoi(strings.Trim(outstr, " \n"))
	if err != nil {
		return 0, err
	}
	return count, nil
}
