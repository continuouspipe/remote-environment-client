package sync

import (
	"github.com/continuouspipe/remote-environment-client/sync/rsync"
)

//fetch all the project files from the pod, or if the filePath is not empty it
//fetch a specific file
type Fetcher interface {
	Fetch(kubeConfigKey string, environment string, pod string, filePath string) error
}

func GetFetcher() Fetcher {
	return rsync.GetRfetch()
}
