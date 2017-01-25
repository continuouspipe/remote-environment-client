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

		var errors string
		if maxFiles < countFilesFolders {
			errorMsg := "kern.maxfiles is %d, which is not enough to watch the amount of project files %d\n" +
				"increase the number of kern.maxfiles with sysctl kern.maxfiles=xyz"
			errors = "\n" + fmt.Sprintf(errorMsg, maxFiles, countFilesFolders) + "\n"
		}
		if maxFilesPerProc < countFilesFolders {
			errorMsg := "kern.maxfilesperproc is %d, which is not enough to watch the amount of project files %d\n" +
				"increase the number of kern.maxfilesperproc with sysctl kern.maxfilesperproc=xyz"
			errors = errors + fmt.Sprintf(errorMsg, maxFilesPerProc, countFilesFolders)
		}
		if len(errors) > 0 {
			return "", fmt.Errorf(errors)
		}

		msg := "kern.maxfiles is %d, which is enough to watch the amount of project files %d.\n" +
			"Note that if you create more than %d files, then you will get an error\n" +
			"and you will have to increase the number of kern.maxfiles with sysctl kern.maxfiles=xyz"
		info = fmt.Sprintf(msg, maxFiles, countFilesFolders, maxFiles-countFilesFolders) + "\n"

		msg = "kern.maxfilesperproc is %d, which is enough to watch the amount of project files %d.\n" +
			"Note that if you create more than %d files, then you will get an error\n" +
			"and you will have to increase the number of kern.maxfilesperproc with sysctl kern.maxfilesperproc=xyz"
		info = info + fmt.Sprintf(msg, maxFiles, countFilesFolders, maxFiles-countFilesFolders)

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
