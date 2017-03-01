package kubectlapi

import (
	"fmt"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/osapi"
	"strings"
)

type KubeCtlInitializer interface {
	Init(environment string) error
}

type kubeCtlClusterSettingsProvider interface {
	settings() (addr string, user string, password string, err error)
}

type KubeCtlInit struct {
	config     config.ConfigProvider
	direct     kubeCtlClusterSettingsProvider
	proxy      kubeCtlClusterSettingsProvider
	kubeConfig kubeCtlConfigProvider
}

func NewKubeCtlInit() *KubeCtlInit {
	i := &KubeCtlInit{}
	i.direct = newKubeCtlDirect()
	i.proxy = newKubeCtlProxy()
	i.config = config.C
	i.kubeConfig = newKubeCtlConfig()
	return i
}

func (i *KubeCtlInit) Init(environment string) error {
	cpKubeProxyEnabled, err := i.config.GetString(config.CpKubeProxyEnabled)
	if err != nil {
		return err
	}

	var addr string
	var user string
	var password string

	if cpKubeProxyEnabled == "true" {
		addr, user, password, err = i.proxy.settings()
	} else {
		addr, user, password, err = i.direct.settings()
	}

	if err != nil {
		return err
	}

	_, err = i.kubeConfig.ConfigSetAuthInfo(environment, user, password)
	if err != nil {
		return err
	}
	_, err = i.kubeConfig.ConfigSetCluster(environment, addr)
	if err != nil {
		return err
	}
	_, err = i.kubeConfig.ConfigSetContext(environment, user)
	if err != nil {
		return err
	}
	return nil
}

type kubeCtlDirect struct {
	config config.ConfigProvider
}

func (i *kubeCtlDirect) settings() (addr string, user string, password string, err error) {
	addr, err = i.config.GetString(config.KubeDirectClusterAddr)
	if err != nil {
		return
	}
	user, err = i.config.GetString(config.KubeDirectClusterUser)
	if err != nil {
		return
	}
	password, err = i.config.GetString(config.KubeDirectClusterPassword)
	if err != nil {
		return
	}
	return
}

func newKubeCtlDirect() *kubeCtlDirect {
	i := &kubeCtlDirect{}
	i.config = config.C
	return i
}

type kubeCtlProxy struct {
	config        config.ConfigProvider
	kubeCtlConfig kubeCtlConfigProvider
}

func (i *kubeCtlProxy) settings() (addr string, user string, password string, err error) {
	flowId, err := i.config.GetString(config.FlowId)
	if err != nil {
		return
	}
	clusterID, err := i.config.GetString(config.ClusterIdentifier)
	if err != nil {
		return
	}
	cpProxyAddr, err := i.config.GetString(config.CpKubeProxyAddr)
	if err != nil {
		return
	}
	addr = fmt.Sprintf("%s/%s/%s/", cpProxyAddr, flowId, clusterID)
	user, err = i.config.GetString(config.Username)
	if err != nil {
		return
	}
	password, err = i.config.GetString(config.ApiKey)
	if err != nil {
		return
	}
	return
}

func newKubeCtlProxy() *kubeCtlProxy {
	i := &kubeCtlProxy{}
	i.config = config.C
	i.kubeCtlConfig = newKubeCtlConfig()
	return i
}

type kubeCtlConfigProvider interface {
	ConfigSetAuthInfo(environment string, username string, password string) (string, error)
	ConfigSetCluster(environment string, clusterAddr string) (string, error)
	ConfigSetContext(environment string, username string) (string, error)
}

type kubeCtlConfig struct{}

func newKubeCtlConfig() *kubeCtlConfig {
	return &kubeCtlConfig{}
}

func (k kubeCtlConfig) ConfigSetAuthInfo(environment string, username string, password string) (string, error) {
	args := []string{
		config.KubeCtlName,
		"config",
		"set-credentials",
		environment + "-" + username,
		"--username=" + username,
		"--password=" + password,
	}
	return osapi.CommandExec(getScmd(), args...)
}

func (k kubeCtlConfig) ConfigSetCluster(environment string, clusterAddr string) (string, error) {

	//avoid double // in address
	clusterAddr = strings.TrimRight(clusterAddr, "/")

	args := []string{
		config.KubeCtlName,
		"config",
		"set-cluster",
		environment,
		fmt.Sprintf("--server=%s", clusterAddr),
		"--insecure-skip-tls-verify=true",
	}
	return osapi.CommandExec(getScmd(), args...)
}

func (k kubeCtlConfig) ConfigSetContext(environment string, username string) (string, error) {
	args := []string{
		config.KubeCtlName,
		"config",
		"set-context",
		environment,
		"--cluster=" + environment,
		"--user=" + environment + "-" + username,
	}
	return osapi.CommandExec(getScmd(), args...)
}
