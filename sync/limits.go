package sync

import (
	"runtime"
	"github.com/continuouspipe/remote-environment-client/osapi"
	"fmt"
	"strconv"
)

type WatchLimitVerifier interface {
	Check(int) (string, error)
}

type WatchLimit struct {
	warningThreshold int
}

func NewWatchLimit() *WatchLimit {
	return &WatchLimit{}
}

// checks if the watch limits set are enough to watch all files and folders in the project
// returns an error if the available watches are less than the required ones
// returns information about available watches if there are enough watches to support the project
func (w WatchLimit) Check(countFilesFolders int) (string, error) {
	var info string
	if runtime.GOOS == "darwin" {
		maxFiles, err := w.getSysctlInt("kern.maxfiles")
		if err != nil {
			return "", err
		}
		maxFilesPerProc, err := w.getSysctlInt("kern.maxfilesperproc")
		if err != nil {
			return "", err
		}

		lsofCount, err := osapi.GetCurrentlyOpenedFilesCount()
		if err != nil {
			return "", err
		}

		suggestedMax := countFilesFolders + lsofCount + 100

		var errors string
		if maxFiles-lsofCount < countFilesFolders {
			errorMsg := "kern.maxfiles is %d, currently open files are %d, which is not enough to watch the amount of project files %d\n" +
				"increase the number of kern.maxfiles with \nsudo sysctl kern.maxfiles=%d\nulimit -S -n %d\n\n"
			errors = "\n" + fmt.Sprintf(errorMsg, maxFiles, lsofCount, countFilesFolders, suggestedMax, suggestedMax)
		}
		if maxFilesPerProc-lsofCount < countFilesFolders {
			errorMsg := "kern.maxfilesperproc is %d, currently open files are %d, which is not enough to watch the amount of project files %d\n" +
				"increase the number of kern.maxfilesperproc with \nsudo sysctl kern.maxfilesperproc=%d\nulimit -S -n %d"
			errors = errors + fmt.Sprintf(errorMsg, maxFilesPerProc, lsofCount, countFilesFolders, suggestedMax, suggestedMax)
		}
		if len(errors) > 0 {
			return "", fmt.Errorf(errors)
		}

		msg := "kern.maxfiles is %d, currently open files are %d, which is enough to watch the amount of project files %d.\n" +
			"Note that if you create more than %d files, then you will get an error\n" +
			"and you will have to increase the number of kern.maxfiles with \nsudo sysctl kern.maxfiles=xyz\nulimit -S -n xyz\n\n"
		info = fmt.Sprintf(msg, maxFiles, lsofCount, countFilesFolders, maxFiles-countFilesFolders)

		msg = "kern.maxfilesperproc is %d, currently open files are %d, which is enough to watch the amount of project files %d.\n" +
			"Note that if you create more than %d files, then you will get an error\n" +
			"and you will have to increase the number of kern.maxfilesperproc with \nsudo sysctl kern.maxfilesperproc=xyz\nulimit -S -n xyz"
		info = info + fmt.Sprintf(msg, maxFiles, lsofCount, countFilesFolders, maxFiles-countFilesFolders)
	}

	return info, nil
}

func (w WatchLimit) getSysctlInt(name string) (int, error) {
	res, err := osapi.CommandExec("sysctl", "-n", name)
	if err != nil {
		return 0, err
	}
	count, err := strconv.Atoi(res)
	if err != nil {
		return 0, err
	}
	return count, nil;
}
