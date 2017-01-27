BINARY=cp-remote-go
VERSION=0.0.1-alpha.5
CONFIG_PKG=github.com/continuouspipe/remote-environment-client/config
LDFLAGS=-ldflags="-X ${CONFIG_PKG}.CurrentVersion=${VERSION}"

build-darwin-amd64:
	mkdir bin 2>/dev/null; true
	env GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o bin/darwin-amd64

build-linux-amd64:
	mkdir bin 2>/dev/null; true
	env GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o bin/linux-amd64

build-windows-amd64:
	mkdir bin 2>/dev/null; true
        env GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o bin/windows-amd64

clean:
	rm -f ${BINARY}
	rm -f update/cktime

update-dep:
	glide up --strip-vendor

# installs the dependencies in the glide.lock file
install-dep:
	glide install --strip-vendor

# install the dependencies for creating a new installation package
install-package-dep:
	go get github.com/sanbornm/go-selfupdate
	cd ${GOPATH}/src/github.com/sanbornm/go-selfupdate && go install
	ln -s ${GOPATH}/bin/go-selfupdate /usr/local/bin/go-selfupdate 2>/dev/null; true

# copy from the github pages repositories all versions into a public/ folder before running this command
# as go-selfupdate needs to find the previous installation in order to generate the binary diff for incremental upgrades
package:
	echo "Note: In order to have the incremental upgrades working properly, rembember to copy all the current releases (including the json) into a local public/ folder"
	go-selfupdate bin/ ${VERSION}
