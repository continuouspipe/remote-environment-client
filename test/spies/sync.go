package spies

//Spy for Syncer interface
type SpySyncer struct {
	//Inherit spy capabilities
	Spy

	//mocked
	sync func(filePaths []string) error
}

func NewSpySyncer() *SpySyncer {
	return &SpySyncer{}
}

//setters for mocks
func (s *SpySyncer) MockSync(mocked func(filePaths []string) error) {
	s.sync = mocked
}

//spies functions
func (s *SpySyncer) Sync(filePaths []string) error {
	args := make(Arguments)
	args["filePaths"] = filePaths

	function := &Function{Name: "Sync", Arguments: args}
	s.calledFunctions = append(s.calledFunctions, *function)
	return s.sync(filePaths)
}

func (s *SpySyncer) SetKubeConfigKey(key string) {
	args := make(Arguments)
	args["key"] = key

	function := &Function{Name: "SetKubeConfigKey", Arguments: args}
	s.calledFunctions = append(s.calledFunctions, *function)
}

func (s *SpySyncer) SetEnvironment(env string) {
	args := make(Arguments)
	args["env"] = env

	function := &Function{Name: "SetEnvironment", Arguments: args}
	s.calledFunctions = append(s.calledFunctions, *function)
}

func (s *SpySyncer) SetPod(pod string) {
	args := make(Arguments)
	args["pod"] = pod

	function := &Function{Name: "SetPod", Arguments: args}
	s.calledFunctions = append(s.calledFunctions, *function)
}

func (s *SpySyncer) SetRemoteProjectPath(remoteProjectPath string) {
	args := make(Arguments)
	args["remoteProjectPath"] = remoteProjectPath

	function := &Function{Name: "SetRemoteProjectPath", Arguments: args}
	s.calledFunctions = append(s.calledFunctions, *function)
}

func (s *SpySyncer) SetIndividualFileSyncThreshold(threshold int) {
	args := make(Arguments)
	args["threshold"] = threshold

	function := &Function{Name: "SetIndividualFileSyncThreshold", Arguments: args}
	s.calledFunctions = append(s.calledFunctions, *function)
}

func (s *SpySyncer) SetVerbose(verbose bool) {
	args := make(Arguments)
	args["verbose"] = verbose

	function := &Function{Name: "SetVerbose", Arguments: args}
	s.calledFunctions = append(s.calledFunctions, *function)
}
