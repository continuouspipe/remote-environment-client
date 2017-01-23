package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"runtime"

	envconfig "github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/util"
	"github.com/sanbornm/go-selfupdate/selfupdate"
)

var selfUpdater = &selfupdate.Updater{
	// Manually update the const, or set it using `go build -ldflags="-X main.VERSION=<newver>" -o cp-remote remote-environment-client/main.go`
	CurrentVersion: envconfig.CurrentVersion,
	// The server hosting `$CmdName/$GOOS-$ARCH.json` which contains the checksum for the binary
	ApiURL: "https://raw.githubusercontent.com/continuouspipe/remote-environment-client/",
	// The server hosting the zip file containing the binary application which is a fallback for the patch method
	BinURL: "https://raw.githubusercontent.com/continuouspipe/remote-environment-client/",
	// The server hosting the binary patch diff for incremental updates
	DiffURL: "https://raw.githubusercontent.com/continuouspipe/remote-environment-client/",
	// The directory created by the app when run which stores the cktime file
	Dir: "update/",
	// The app name which is appended to the ApiURL to look for an update
	CmdName: "gh-pages",
}

func CheckForLatestVersion() error {
	err := fetchInfo(selfUpdater)
	if err != nil {
		return err
	}

	if selfUpdater.Info.Version == selfUpdater.CurrentVersion {
		return nil
	}

	q := &util.QuestionPrompt{}
	res := q.RepeatIfEmpty(fmt.Sprintf("New version available: New version %s is available. Do you want to update now (yes/no)", selfUpdater.Info.Version))

	if res == "no" {
		return nil
	}

	selfUpdater.BackgroundRun()
	return nil
}

//borrowed from the selfupdate package, downloads the json information and popualtes the updater.Info field
func fetchInfo(u *selfupdate.Updater) error {
	platform := runtime.GOOS + "-" + runtime.GOARCH
	r, err := fetch(u, u.ApiURL+url.QueryEscape(u.CmdName)+"/"+url.QueryEscape(platform)+".json")
	if err != nil {
		return err
	}
	defer r.Close()
	err = json.NewDecoder(r).Decode(&u.Info)
	if err != nil {
		return err
	}

	return nil
}

func fetch(u *selfupdate.Updater, url string) (io.ReadCloser, error) {
	defaultHTTPRequester := &selfupdate.HTTPRequester{}

	if u.Requester == nil {
		return defaultHTTPRequester.Fetch(url)
	}

	readCloser, err := u.Requester.Fetch(url)
	if err != nil {
		return nil, err
	}

	if readCloser == nil {
		return nil, fmt.Errorf("Fetch was expected to return non-nil ReadCloser")
	}

	return readCloser, nil
}
