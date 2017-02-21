package sync

import (
	"github.com/continuouspipe/remote-environment-client/sync/rsync"
)

//fetch all the project files from the pod, or if the filePath is not empty it
//fetch a specific file
type Fetcher interface {
	Fetch(string) error
	SetKubeConfigKey(string)
	SetEnvironment(string)
	SetPod(string)
	SetRemoteProjectPath(string)
}

func GetFetcher() Fetcher {
	return rsync.GetRfetch()
}
