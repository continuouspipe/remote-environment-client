package update

import (
	"github.com/continuouspipe/remote-environment-client/cplogs"
	"github.com/sanbornm/go-selfupdate/selfupdate"
	"io"
	"strings"
)

type HttpRequesterWrapper struct {
	defaultRequester *selfupdate.HTTPRequester
}

func NewHttpRequesterWrapper() *HttpRequesterWrapper {
	w := &HttpRequesterWrapper{}
	w.defaultRequester = &selfupdate.HTTPRequester{}
	return w
}

func (r *HttpRequesterWrapper) Fetch(url string) (io.ReadCloser, error) {
	body, err := r.defaultRequester.Fetch(url)

	if err != nil {
		//if the binary diff was missing suppress the error and leave the fallback to kick-in
		if strings.Contains(err.Error(), "bad http status from") {
			cplogs.V(5).Infoln(err.Error())
			cplogs.Flush()
			return body, nil
		}
	}
	cplogs.Flush()
	return body, err
}
