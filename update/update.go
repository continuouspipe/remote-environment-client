package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime"

	"github.com/continuouspipe/remote-environment-client/config"
	cperrors "github.com/continuouspipe/remote-environment-client/errors"
	"github.com/continuouspipe/remote-environment-client/util"
	"github.com/pkg/errors"
	"github.com/sanbornm/go-selfupdate/selfupdate"
)

func NewSelfUpdater() *selfupdate.Updater {
	cfg := config.NewConfig()
	awsBucketAddr, _ := cfg.GetString(config.AwsS3BucketAddr)
	return &selfupdate.Updater{
		// Manually update the const, or set it using `go build -ldflags="-X main.VERSION=<newver>" -o cp-remote remote-environment-client/main.go`
		CurrentVersion: config.CurrentVersion,
		// The server hosting `$CmdName/$GOOS-$ARCH.json` which contains the checksum for the binary
		ApiURL: awsBucketAddr,
		// The server hosting the zip file containing the binary application which is a fallback for the patch method
		BinURL: awsBucketAddr,
		// The server hosting the binary patch diff for incremental updates
		DiffURL: awsBucketAddr,
		// Check for update regardless of cktime timestamp
		ForceCheck: true,
		// The app name which is appended to the ApiURL to look for an update
		CmdName: "downloads",
	}
}

// CheckForLatestVersion looks is there is a new version available, if there is one it will ask the user if he would like to upgrade
func CheckForLatestVersion() error {
	selfUpdater := NewSelfUpdater()

	err := fetchInfo(selfUpdater)
	if err != nil {
		return errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, "error when fetching info").String())
	}

	if selfUpdater.Info.Version == selfUpdater.CurrentVersion {
		return nil
	}

	q := &util.QuestionPrompt{}
	res := q.RepeatIfEmpty(fmt.Sprintf("New version available: New version %s is available. Do you want to update now (yes/no)", selfUpdater.Info.Version))

	if res == "no" {
		return nil
	}

	fmt.Println("Upgrade in progress...")
	selfUpdater.Requester = NewHttpRequesterWrapper()
	return selfUpdater.BackgroundRun()
}

//borrowed from the selfupdate package, downloads the json information and popualtes the updater.Info field
func fetchInfo(u *selfupdate.Updater) error {
	platform := runtime.GOOS + "-" + runtime.GOARCH
	urlPath := u.ApiURL + url.QueryEscape(u.CmdName) + "/" + url.QueryEscape(platform) + ".json"
	r, err := fetch(u, urlPath)
	if err != nil {
		return errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, fmt.Sprintf("failed fetching json manifesto from url %s", urlPath)).String())
	}
	defer r.Close()
	err = json.NewDecoder(r).Decode(&u.Info)
	if err != nil {
		return errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, "error decoding json response").String())
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
		return nil, errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, "requester failed to fetch").String())
	}

	if readCloser == nil {
		return nil, errors.New(cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, "Fetch was expected to return non-nil ReadCloser").String())
	}

	return readCloser, nil
}
