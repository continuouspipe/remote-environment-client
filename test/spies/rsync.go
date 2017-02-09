package spies

//Spy for RsyncFetch
type SpyRsyncFetch struct {
	Spy
	fetch func() error
}

func NewSpyRsyncFetch() *SpyRsyncFetch {
	return &SpyRsyncFetch{}
}

func (r *SpyRsyncFetch) Fetch(kubeConfigKey string, environment string, pod string, filePath string) error {
	args := make(Arguments)
	args["kubeConfigKey"] = kubeConfigKey
	args["environment"] = environment
	args["pod"] = pod
	args["filePath"] = filePath
	function := &Function{Name: "Fetch", Arguments: args}
	r.calledFunctions = append(r.calledFunctions, *function)
	return r.fetch()
}

func (r *SpyRsyncFetch) MockFetch(mocked func() error) {
	r.fetch = mocked
}
