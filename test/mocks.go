package test

import "k8s.io/client-go/pkg/api/v1"

type MockConfigReader struct {
	getString func(string) string
}

func GetMockConfigReader() *MockConfigReader {
	return &MockConfigReader{}
}

func (m *MockConfigReader) GetString(key string) string {
	return m.getString(key)
}

func (m *MockConfigReader) MockGetString(mocked func(string) string) {
	m.getString = mocked
}

type MockPodsFinder struct {
	findAll func(kubeConfigKey string, environment string) (*v1.PodList, error)
}

func GetMockPodsFinder() *MockPodsFinder {
	return &MockPodsFinder{}
}

func (m *MockPodsFinder) FindAll(kubeConfigKey string, environment string) (*v1.PodList, error) {
	return m.findAll(kubeConfigKey, environment)
}

func (m *MockPodsFinder) MockFindAll(mocked func(kubeConfigKey string, environment string) (*v1.PodList, error)) {
	m.findAll = mocked
}

type MockPodsFilter struct {
	byService func(podList *v1.PodList, service string) (*v1.Pod, error)
}

func GetMockPodsFilter() *MockPodsFilter {
	return &MockPodsFilter{}
}

func (m *MockPodsFilter) ByService(podList *v1.PodList, service string) (*v1.Pod, error) {
	return m.byService(podList, service)
}

func (m *MockPodsFilter) MockByService(mocked func(podList *v1.PodList, service string) (*v1.Pod, error)) {
	m.byService = mocked
}

type Arguments map[string]interface{}

//Function is a struct where you can set the name and add a slice Arguments ([]Argument) for each call
type Function struct {
	Name      string
	Arguments Arguments
}

type MockLocalExecutor struct {
	calledFunctions []Function
	commandExec     func() string
}

func GetMockLocalExecutor() *MockLocalExecutor {
	return &MockLocalExecutor{}
}

func (m *MockLocalExecutor) FirstCallsFor(functionName string) *Function {
	for _, call := range m.calledFunctions {
		if call.Name == functionName {
			return &call
		}
	}
	return nil
}

func (m *MockLocalExecutor) CallsCountFor(functionName string) int {
	count := 0
	for _, call := range m.calledFunctions {
		if call.Name != functionName {
			continue
		}
		count++
	}
	return count
}

func (m *MockLocalExecutor) SysCallExec(kubeConfigKey string, environment string, pod string, execCmdArgs ...string) {
	args := make(Arguments)
	args["kubeConfigKey"] = kubeConfigKey
	args["environment"] = environment
	args["pod"] = pod
	args["execCmdArgs"] = execCmdArgs

	function := &Function{Name: "SysCallExec", Arguments: args}
	m.calledFunctions = append(m.calledFunctions, *function)
}

func (m *MockLocalExecutor) CommandExec(kubeConfigKey string, environment string, pod string, execCmdArgs ...string) string {
	args := make(Arguments)
	args["kubeConfigKey"] = kubeConfigKey
	args["environment"] = environment
	args["pod"] = pod
	args["execCmdArgs"] = execCmdArgs

	function := &Function{Name: "CommandExec", Arguments: args}
	m.calledFunctions = append(m.calledFunctions, *function)

	return m.commandExec()
}

func (m *MockLocalExecutor) MockCommandExec(mocked func() string) {
	m.commandExec = mocked
}
