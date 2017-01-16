BINARY=cp-remote-go
VERSION=0.0.1-alpha.2
CONFIG_PKG=github.com/continuouspipe/remote-environment-client/config
LDFLAGS=-ldflags="-X ${CONFIG_PKG}.CurrentVersion=${VERSION}"

build:
	go build ${LDFLAGS} -o ${BINARY}

clean:
	rm -f ${BINARY}
	rm -f update/cktime

install-dep:
	glide install --strip-vendor --strip-vcs

package:
	go-selfupdate ${BINARY} ${VERSION}
