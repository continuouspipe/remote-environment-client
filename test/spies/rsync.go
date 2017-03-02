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
type SpyRsyncPush struct {
	Spy
	push func() error
}

func NewSpyRsyncPush() *SpyRsyncPush {
	return &SpyRsyncPush{}
}

func (r *SpyRsyncPush) SetKubeConfigKey(kubeConfigKey string) {
	args := make(Arguments)
	args["kubeConfigKey"] = kubeConfigKey
	function := &Function{Name: "SetKubeConfigKey", Arguments: args}
	r.calledFunctions = append(r.calledFunctions, *function)
}
func (r *SpyRsyncPush) SetEnvironment(environment string) {
	args := make(Arguments)
	args["environment"] = environment
	function := &Function{Name: "SetEnvironment", Arguments: args}
	r.calledFunctions = append(r.calledFunctions, *function)
}
func (r *SpyRsyncPush) SetPod(pod string) {
	args := make(Arguments)
	args["pod"] = pod
	function := &Function{Name: "SetPod", Arguments: args}
	r.calledFunctions = append(r.calledFunctions, *function)
}
func (r *SpyRsyncPush) SetRemoteProjectPath(remoteProjectPath string) {
	args := make(Arguments)
	args["remoteProjectPath"] = remoteProjectPath
	function := &Function{Name: "SetRemoteProjectPath", Arguments: args}
	r.calledFunctions = append(r.calledFunctions, *function)
}

func (r *SpyRsyncPush) Push(filePath string) error {
	args := make(Arguments)
	args["filePath"] = filePath
	function := &Function{Name: "Push", Arguments: args}
	r.calledFunctions = append(r.calledFunctions, *function)
	return r.push()
}

func (r *SpyRsyncPush) MockPush(mocked func() error) {
	r.push = mocked
}
