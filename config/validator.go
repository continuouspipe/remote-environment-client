package config

type Validator interface {
	Validate(Reader) (n int, missing []string)
}

type MandatoryChecker struct {
	settings []string
}

//TODO: Change to use the local/global config settings
func NewMandatoryChecker() *MandatoryChecker {
	checker := &MandatoryChecker{}
	checker.settings = []string{
		Username,
		Password,
		Team,
		ClusterId,
		ProjectKey,
		RemoteBranch,
		RemoteName,
		Service,
		KubeConfigKey}
	return checker
}

//takes the application config from a config reader and checks that all the required fields are populated
func (checker *MandatoryChecker) Validate(configReader Reader) (n int, missing []string) {
	for _, setting := range checker.settings {
		if settingValue := configReader.GetString(setting); settingValue == "" {
			missing = append(missing, setting)
		}
	}
	return len(missing), missing
}
