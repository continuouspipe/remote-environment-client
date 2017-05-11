//Package cplogs This file contains the code that handle logging information that get sent to an external server
//We send warning alerts when something that we do not expect happens
//We send operational metrics for each command such as, command name, arguments, duration, status etc..
package remote

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"time"

	"bytes"

	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	cperrors "github.com/continuouspipe/remote-environment-client/errors"
	cphttp "github.com/continuouspipe/remote-environment-client/http"
	"github.com/continuouspipe/remote-environment-client/session"
	"github.com/pkg/errors"
)

//ErrorFailedToSendDataToLoggingAPI is used as a log message when the post request to the logging api had failed
const ErrorFailedToSendDataToLoggingAPI = "failed to send the post request to the logging API"

const errorFailedToRetrievedLoggingAPIURL = "failed to retrieve the logging api url"

//RemoteCommand holds the data that is logged when a command terminates
type RemoteCommand struct {
	//to be set once the command terminates
	Duration time.Duration       `json:"duration"`
	Status   RemoteCommandStatus `json:"status"`

	//to be set at the beginning of the command execution
	Command        string                `json:"command"`
	Arguments      []string              `json:"arguments"`
	OsArch         string                `json:"os_arch"`
	ToolVersion    string                `json:"tool_version"`
	ConfigSettings RemoteCommandSettings `json:"config_settings"`

	//to be set at the beginning of the command execution, only for fetch, push, watch
	IgnoreFileContent string `json:"ignore_file_content"`

	//to be set at the beginning of the command execution, only for fetch
	IgnoreFetchFileContent string `json:"ignore_fetch_file_content"`
}

//RemoteCommandSettings contains a subset of the local configuration data that we want to send along with logging information
type RemoteCommandSettings struct {
	FlowID                string `json:"flow-id"`
	ClusterIdentifier     string `json:"cluster-identifier"`
	KubeEnvironmentName   string `json:"kube-environment-name"`
	RemoteName            string `json:"remote-name"`
	RemoteBranch          string `json:"remote-branch"`
	Service               string `json:"service"`
	RemoteEnvironmentID   string `json:"remote-environment-id"`
	InitStatus            string `json:"init-status"`
	CpKubeProxyEnabled    string `json:"kube-proxy-enabled"`
	KubeDirectClusterAddr string `json:"kube-direct-cluster-addr"`
	KubeDirectClusterUser string `json:"kube-direct-cluster-user"`
}

//RemoteCommandStatus groups status information about the remote command execution
type RemoteCommandStatus struct {
	//An integer that represents the status. Should be between 100 and 999. Same as HTTP codes, the first digit categorises the event:
	//2xx: Success
	//3xx: Worked, but.. not sure if it should have happened
	//4xx: Something went wrong, but that’s very likely to be user’s fault
	//5xx: Something went wrong, very likely to be system’s fault
	Code int `json:"code"`

	//The reason describing why the event is a success or a failure.
	Reason string `json:"reason"`

	//The full error stack
	Stack string `json:"error_stack"`

	//A unique id which can be used to find any messages sent to Sentry or in the logs
	DebugIdentifier string `json:"debug_identifier"`
}

//NewRemoteCommand create a new remote command struct for the given command and arguments
func NewRemoteCommand(cmd string, args []string) *RemoteCommand {
	return &RemoteCommand{
		Command:     cmd,
		Arguments:   args,
		OsArch:      fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH),
		ToolVersion: config.CurrentVersion,
		ConfigSettings: RemoteCommandSettings{
			FlowID:                config.C.GetStringQ(config.FlowId),
			ClusterIdentifier:     config.C.GetStringQ(config.ClusterIdentifier),
			KubeEnvironmentName:   config.C.GetStringQ(config.KubeEnvironmentName),
			RemoteName:            config.C.GetStringQ(config.RemoteName),
			RemoteBranch:          config.C.GetStringQ(config.RemoteBranch),
			Service:               config.C.GetStringQ(config.Service),
			RemoteEnvironmentID:   config.C.GetStringQ(config.RemoteEnvironmentId),
			InitStatus:            config.C.GetStringQ(config.InitStatus),
			CpKubeProxyEnabled:    config.C.GetStringQ(config.CpKubeProxyEnabled),
			KubeDirectClusterAddr: config.C.GetStringQ(config.KubeDirectClusterAddr),
			KubeDirectClusterUser: config.C.GetStringQ(config.KubeDirectClusterUser),
		},
	}
}

//EndedOk calls Ended setting the http status OK status
func (rc *RemoteCommand) EndedOk(cmdSession session.CommandSession) *RemoteCommand {
	rc.Ended(http.StatusOK, "", "", cmdSession)
	return rc
}

//Ended indicates that the remote command has terminated its execution sets the duration and the command status
func (rc *RemoteCommand) Ended(code int, reason string, stack string, cmdSession session.CommandSession) *RemoteCommand {
	rc.Status = RemoteCommandStatus{
		Code:            code,
		Reason:          reason,
		Stack:           stack,
		DebugIdentifier: cmdSession.SessionID,
	}
	rc.Duration = cmdSession.Duration()
	return rc
}

//RemoteCommandSender holds the dependencies required for the RemoteCommandSender
type RemoteCommandSender struct{}

//NewRemoteCommandSender ctor for RemoteCommandSender
func NewRemoteCommandSender() *RemoteCommandSender {
	return &RemoteCommandSender{}
}

//Send sends a RemoteCommand to the log proxy url
func (s RemoteCommandSender) Send(rc RemoteCommand) error {
	u, err := GetLogProxyAddr()
	if err != nil {
		return errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, errorFailedToRetrievedLoggingAPIURL).String())
	}
	u.Path = "/logs"

	reqBodyJSON, err := json.Marshal(rc)
	if err != nil {
		msg := fmt.Sprintf(cphttp.ErrorCreatingJSONRequest, reqBodyJSON)
		cplogs.V(4).Infof(msg)
		return errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, msg).String())
	}

	req, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewReader(reqBodyJSON))
	if err != nil {
		return errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, cphttp.ErrorFailedToCreatePostRequest).String())
	}
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	_, err = cphttp.GetResponseBody(client, req)
	if err != nil {
		cplogs.V(4).Infof(cphttp.ErrorFailedToGetResponseBody, u.String())
		cplogs.Flush()
		return errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, cphttp.ErrorFailedToGetResponseBody).String())
	}
	return nil
}

//GetLogProxyAddr gets the api address for the cp logging proxy
func GetLogProxyAddr() (*url.URL, error) {
	addr, err := config.C.GetString(config.CpLogProxyAddr)
	if err != nil {
		return nil, err
	}
	u, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}
	return u, nil
}
