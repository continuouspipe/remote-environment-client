package rsync

//use rsync to sync the files specified in filePaths. When filePaths is an empty slice, it syncs all project files
type RsyncSyncer interface {
	Sync(filePaths []string) error
	SetKubeConfigKey(string)
	SetEnvironment(string)
	SetPod(string)
	SetIndividualFileSyncThreshold(int)
}

//this is initialised by either by rsyncrsh or rsyncdaemon depending on the build constrains
var Rsync RsyncSyncer
