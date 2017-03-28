package spies

//Spy for RsyncFetch
type SpyRsyncFetch struct {
	Spy
	fetch func() error
}

func NewSpyRsyncFetch() *SpyRsyncFetch {
	return &SpyRsyncFetch{}
}

func (r *SpyRsyncFetch) SetKubeConfigKey(kubeConfigKey string) {
	args := make(Arguments)
	args["kubeConfigKey"] = kubeConfigKey
	function := &Function{Name: "SetKubeConfigKey", Arguments: args}
	r.calledFunctions = append(r.calledFunctions, *function)
}

func (r *SpyRsyncFetch) SetEnvironment(environment string) {
	args := make(Arguments)
	args["environment"] = environment
	function := &Function{Name: "SetEnvironment", Arguments: args}
	r.calledFunctions = append(r.calledFunctions, *function)
}

func (r *SpyRsyncFetch) SetPod(pod string) {
	args := make(Arguments)
	args["pod"] = pod
	function := &Function{Name: "SetPod", Arguments: args}
	r.calledFunctions = append(r.calledFunctions, *function)
}

func (r *SpyRsyncFetch) SetRemoteProjectPath(remoteProjectPath string) {
	args := make(Arguments)
	args["remoteProjectPath"] = remoteProjectPath
	function := &Function{Name: "SetRemoteProjectPath", Arguments: args}
	r.calledFunctions = append(r.calledFunctions, *function)
}

func (r *SpyRsyncFetch) SetVerbose(verbose bool) {
	args := make(Arguments)
	args["verbose"] = verbose
	function := &Function{Name: "SetVerbose", Arguments: args}
	r.calledFunctions = append(r.calledFunctions, *function)
}

func (r *SpyRsyncFetch) Fetch(filePath string) error {
	args := make(Arguments)
	args["filePath"] = filePath
	function := &Function{Name: "Fetch", Arguments: args}
	r.calledFunctions = append(r.calledFunctions, *function)
	return r.fetch()
}

func (r *SpyRsyncFetch) MockFetch(mocked func() error) {
	r.fetch = mocked
}

//Spy for RsyncPush
type SpyRsyncSyncer struct {
	Spy
	sync func() error
}

func NewSpyRsyncSyncer() *SpyRsyncSyncer {
	return &SpyRsyncSyncer{}
}

func (r *SpyRsyncSyncer) SetKubeConfigKey(kubeConfigKey string) {
	args := make(Arguments)
	args["kubeConfigKey"] = kubeConfigKey
	function := &Function{Name: "SetKubeConfigKey", Arguments: args}
	r.calledFunctions = append(r.calledFunctions, *function)
}

func (r *SpyRsyncSyncer) SetEnvironment(environment string) {
	args := make(Arguments)
	args["environment"] = environment
	function := &Function{Name: "SetEnvironment", Arguments: args}
	r.calledFunctions = append(r.calledFunctions, *function)
}

func (r *SpyRsyncSyncer) SetPod(pod string) {
	args := make(Arguments)
	args["pod"] = pod
	function := &Function{Name: "SetPod", Arguments: args}
	r.calledFunctions = append(r.calledFunctions, *function)
}

func (r *SpyRsyncSyncer) SetRemoteProjectPath(remoteProjectPath string) {
	args := make(Arguments)
	args["remoteProjectPath"] = remoteProjectPath
	function := &Function{Name: "SetRemoteProjectPath", Arguments: args}
	r.calledFunctions = append(r.calledFunctions, *function)
}

func (r *SpyRsyncSyncer) SetIndividualFileSyncThreshold(individualFileSyncThreshold int) {
	args := make(Arguments)
	args["individualFileSyncThreshold"] = individualFileSyncThreshold
	function := &Function{Name: "SetIndividualFileSyncThreshold", Arguments: args}
	r.calledFunctions = append(r.calledFunctions, *function)
}

func (r *SpyRsyncSyncer) SetVerbose(verbose bool) {
	args := make(Arguments)
	args["verbose"] = verbose
	function := &Function{Name: "SetVerbose", Arguments: args}
	r.calledFunctions = append(r.calledFunctions, *function)
}

func (r *SpyRsyncSyncer) Sync(filePaths []string) error {
	args := make(Arguments)
	args["filePaths"] = filePaths
	function := &Function{Name: "Sync", Arguments: args}
	r.calledFunctions = append(r.calledFunctions, *function)
	return r.sync()
}

func (r *SpyRsyncSyncer) MockSync(mocked func() error) {
	r.sync = mocked
}
