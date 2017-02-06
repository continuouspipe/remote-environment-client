// +build windows

package rsync

func init() {
	RsyncDaemon = NewRSyncDaemon()
}

type RSyncDaemon struct{}

func NewRSyncDaemon() *RSyncDaemon {
	return &RSyncDaemon{}
}

func (o *RSyncDaemon) Sync(filePaths []string) error {
	//TODO: Implement for windows
	return nil
}
func (o *RSyncDaemon) SetKubeConfigKey(string) {

}
func (o *RSyncDaemon) SetEnvironment(string) {

}
func (o *RSyncDaemon) SetPod(string) {

}
func (o *RSyncDaemon) SetIndividualFileSyncThreshold(int) {

}
