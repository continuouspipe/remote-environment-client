BINARY=cp-remote-go
VERSION=0.0.1-alpha.1
CONFIG_PKG=github.com/continuouspipe/remote-environment-client/config
LDFLAGS=-ldflags="-X ${CONFIG_PKG}.CurrentVersion=${VERSION}"

build:
	go build ${LDFLAGS} -o ${BINARY}

clean:
	rm -f ${BINARY}

package:
	go-selfupdate ${BINARY} ${VERSION}
