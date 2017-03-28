package sync

import (
	"github.com/continuouspipe/remote-environment-client/sync/rsync"
	"github.com/continuouspipe/remote-environment-client/sync/options"
)

//fetch all the project files from the pod, or if the filePath is not empty it
//fetch a specific file
type Fetcher interface {
	Fetch(string) error
	SetOptions(syncOptions options.SyncOptions)
}

func GetFetcher() Fetcher {
	return rsync.GetRfetch()
}
