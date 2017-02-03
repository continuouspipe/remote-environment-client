package rsync

import "runtime"

//use rsync to fetch all the project files from the pod, or if the filePath is not empty it
//fetch a specific file
type RsyncFetcher interface {
	Fetch(kubeConfigKey string, environment string, pod string, filePath string) error
}

var RfetchRsh RsyncFetcher
var RfetchDaemon RsyncFetcher

func GetRfetch() RsyncFetcher {
	if runtime.GOOS == "windows" {
		return RfetchDaemon
	}
	return RfetchRsh
}