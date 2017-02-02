package rsync

//use rsync to fetch all the project files from the pod, or if the filePath is not empty it
//fetch a specific file
type RsyncFetcher interface {
	Fetch(kubeConfigKey string, environment string, pod string, filePath string) error
}

//this is initialised by either by fetchrsh or fetchdaemon depending on the build constrains
var Rfetch RsyncFetcher
