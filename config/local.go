package config

type LocalConfig struct {
	settings []Setting
	viperWrapper
}

func NewLocalConfig() *LocalConfig {
	local := &LocalConfig{}
	local.settings = []Setting{
		{"project", "", true},                //CP project name (previously Team Name)
		{"flow", "", true},                   //CP flow Name
		{"cluster-identifier", "", true},     //CP cluster Identifier
		{"project-key", "", true},            //CP project key
		{"remote-name", "origin", true},      //Github remote name (origin by default)
		{"remote-branch", "", true},          //Git name of the git branch used for the remote environment
		{"service", "web", true},             //Kubernetes service name for the commands like (watch, bash, fetch and resync)
		{"kubernetes-config-key", "", true},  //Kubernetes configuration file key
		{"anybar-port", "", false},           //AnyBar port number
		{"keen-write-key", "", false},        //Keen.io write key
		{"keen-project-id", "", false},       //Keen.io project id
		{"keen-event-collection", "", false}, //Keen.io event collection
	}
	return local
}

/*func GetEnvironment(projectKey string, remoteBranch string) string {
	environment := strings.Replace(remoteBranch, string(filepath.Separator), "-", -1)
	environment = strings.Replace(environment, "\\", "-", -1)
	environment = projectKey + "-" + environment
	return environment
}*/
