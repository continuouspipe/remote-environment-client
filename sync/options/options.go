package options

type SyncOptions struct {
	KubeConfigKey, Environment, Pod, RemoteProjectPath string
	IndividualFileSyncThreshold                        int
	Verbose, DryRun, Delete                            bool
}
