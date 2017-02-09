package spies

//Spy for Syncer interface
type SpySyncer struct {
	//Inherit spy capabilities
	Spy

	//from interface
	Sync                           func(filePaths []string) error
	SetKubeConfigKey               func(string)
	SetEnvironment                 func(string)
	SetPod                         func(string)
	SetIndividualFileSyncThreshold func(int)

	//setters for mocks
	MockSync                           func(mocked func(filePaths []string) error)
	MockSetKubeConfigKey               func(mocked func(string))
	MockSetEnvironment                 func(mocked func(string))
	MockSetPod                         func(mocked func(string))
	MockSetIndividualFileSyncThreshold func(mocked func(int))

	//mocked
	sync                           func(filePaths []string) error
	setKubeConfigKey               func(string)
	setEnvironment                 func(string)
	setPod                         func(string)
	setIndividualFileSyncThreshold func(int)
}

func NewSpySyncer() *SpySyncer {
	return &SpySyncer{}
}
