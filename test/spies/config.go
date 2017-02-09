package spies

import (
	"github.com/continuouspipe/remote-environment-client/config"
)

//Spy for YamlWriter
type SpyYamlWriter struct {
	Spy
	save func() bool
}

func NewSpyYamlWriter() *SpyYamlWriter {
	return &SpyYamlWriter{}
}

func (m *SpyYamlWriter) Save(settings *config.ApplicationSettings) bool {
	copySettings := *settings

	args := make(Arguments)
	args["settings"] = &copySettings

	function := &Function{Name: "Save", Arguments: args}
	m.calledFunctions = append(m.calledFunctions, *function)
	return m.save()
}

func (m *SpyYamlWriter) MockSave(mocked func() bool) {
	m.save = mocked
}
