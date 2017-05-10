package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

var counter int = 0

func main() {
	http.HandleFunc("/api/remote-environment/123456", ServeGetRemoteEnvironment)
	http.HandleFunc("/api/remote-environment/123456/", ServeGetRemoteEnvironment)
	http.HandleFunc("/api/remote-environment/123456/build", ServeGetRemoteEnvironment)
	http.HandleFunc("/api/remote-environment/123456/build/", ServeGetRemoteEnvironment)
	err := http.ListenAndServe("localhost:31986", nil)
	if err != nil {
		panic(err)
	}
}

type ApiRemoteEnvironment struct {
	Status              string `json:"status"`
	ModifiedAt          string `json:"modified_at"`
	RemoteEnvironmentId string `json:"remote_environment_id"`
	KubeEnvironmentName string `json:"kubernetes_environment_name"`
	ClusterIdentifier   string `json:"cluster_identifier"`
	AnyBarPort          string `json:"any_bar_port"`
}

func ServeGetRemoteEnvironment(w http.ResponseWriter, r *http.Request) {
	counter = counter + 1
	var status string

	switch counter {
	case 1:
		status = "NotStarted"
	case 2:
		status = "Building"
	default:
		status = "Ok"
	}

	w.Header().Add("Content-Type", "application/json")

	t := time.Now()

	apiRemoteEnvironment := &ApiRemoteEnvironment{
		status,
		t.Format(time.RFC3339),
		"123456",
		"1268cc54-b265-11e6-b835-0c360641bb54-remote-alessandrozucca",
		"strava-de-france",
		""}
	p, _ := json.Marshal(apiRemoteEnvironment)
	fmt.Fprint(w, string(p))
}
