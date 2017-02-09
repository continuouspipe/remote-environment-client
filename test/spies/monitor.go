package spies

import (
	"github.com/continuouspipe/remote-environment-client/sync/monitor"
	"time"
)

//Spy for OsDirectoryMonitor
type SpyOsDirectoryMonitor struct {
	Spy
	anyEventCall func(directory string, observer monitor.EventsObserver) error
}

func NewSpyOsDirectoryMonitor() *SpyOsDirectoryMonitor {
	return &SpyOsDirectoryMonitor{}
}

func (m *SpyOsDirectoryMonitor) AnyEventCall(directory string, observer monitor.EventsObserver) error {
	args := make(Arguments)
	args["directory"] = directory
	args["observer"] = observer

	function := &Function{Name: "AnyEventCall", Arguments: args}
	m.calledFunctions = append(m.calledFunctions, *function)

	return m.anyEventCall(directory, observer)
}

func (m *SpyOsDirectoryMonitor) MockAnyEventCall(mocked func(directory string, observer monitor.EventsObserver) error) {
	m.anyEventCall = mocked
}

func (m *SpyOsDirectoryMonitor) SetExclusions(exclusion monitor.ExclusionProvider) {
	args := make(Arguments)
	args["exclusion"] = exclusion

	function := &Function{Name: "SetExclusions", Arguments: args}
	m.calledFunctions = append(m.calledFunctions, *function)
}

func (m *SpyOsDirectoryMonitor) SetLatency(latency time.Duration) {
	args := make(Arguments)
	args["latency"] = latency

	function := &Function{Name: "SetLatency", Arguments: args}
	m.calledFunctions = append(m.calledFunctions, *function)
}
