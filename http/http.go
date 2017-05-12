package http

import (
	"fmt"
	"io/ioutil"
	"net/http"

	cperrors "github.com/continuouspipe/remote-environment-client/errors"
	"github.com/pkg/errors"
)

const ErrorResponseStatusCodeUnsuccessful = "failed to get response body, status: %d, url: %s"
const ErrorFailedToCreateGetRequest = "failed to create a get request"
const ErrorFailedToCreatePostRequest = "failed to create a post request"
const ErrorFailedToCreateDeleteRequest = "failed to create a delete request"
const ErrorParsingJSONResponse = "failed to unparse json response body %s"
const ErrorCreatingJSONRequest = "failed to create json request from struct %v"
const ErrorFailedToGetResponseBody = "failed to get the response body: %s"

func GetResponseBody(client *http.Client, req *http.Request) ([]byte, error) {
	res, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, err.Error()).String())
	}
	defer res.Body.Close()
	if res.StatusCode < 200 && res.StatusCode > 202 {
		errStr := fmt.Sprintf(ErrorResponseStatusCodeUnsuccessful, res.StatusCode, req.URL.String())
		return nil, errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, errStr).String())
	}
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, err.Error()).String())
	}
	return resBody, nil
}
