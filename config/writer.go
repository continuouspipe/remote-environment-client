package config

import (
	"text/template"
	"os"
)

var AppName = os.Args[0]

const KubeCtlName = "kubectl"

type Reader interface {
	GetString(key string) string
}

func (s ApplicationSettings) GetString(key string) string {
	return s.viper.GetString(key)
}

type Writer interface {
	Save(filePath string, t *template.Template) bool
}
type YamlWriter struct{}

func NewYamlWriter() *YamlWriter {
	return &YamlWriter{}
}

//save on the settings file the config data
func (writer YamlWriter) Save(filePath string, t *template.Template) bool {
	f, err := os.Create(filePath)
	if err != nil {
		return false
	}

	err = t.Execute(f, config)
	return err == nil
}

