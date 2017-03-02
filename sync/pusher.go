package sync

//fetch all the project files from the pod, or if the filePath is not empty it
//fetch a specific file
type Pusher interface {
	Push(string) error
	SetKubeConfigKey(string)
	SetEnvironment(string)
	SetPod(string)
	SetRemoteProjectPath(string)
}

func GetPusher() Pusher {
	return nil
}
