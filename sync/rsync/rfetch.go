package rsync

import (
	"runtime"
	"github.com/continuouspipe/remote-environment-client/sync/options"
)

//use rsync to fetch all the project files from the pod, or if the filePath is not empty it
//fetch a specific file
type RsyncFetcher interface {
	Fetch(string) error
	SetOptions(syncOptions options.SyncOptions)
}

var RfetchRsh RsyncFetcher
var RfetchDaemon RsyncFetcher

func GetRfetch() RsyncFetcher {
	if runtime.GOOS == "windows" {
		return RfetchDaemon
	}
	return RfetchRsh
}
