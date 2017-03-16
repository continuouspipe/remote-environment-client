# In order release a new version, checkout from the github pages at continuouspipe/remote-environment-client/gh-pages
# all the GOOS-GOARCH.json files e.g.(darwin-amd64.json, windows-386.json, and so on..) along with the previous
# 2 releases folders e.g.(0.0.1, 0.0.2 if you are releasing 0.0.3) and copy this in a folder called "public"
#
# Then cross-compile for all supported goos-goarch:
#
# make build BUILDOS=darwin BUILDARCH=amd64
# make build BUILDOS=windows BUILDARCH=amd64
# make build BUILDOS=windows BUILDARCH=386
# make build BUILDOS=linux BUILDARCH=amd64
# make build BUILDOS=linux BUILDARCH=386
#
#
# Auto-Upgrade Section:
#
# make build will put the binaries into the bin/ folder.
# After this is done run "make package" which will put the new binaries into public/ along with the binary diffs
# in order to have a quicker upgrade from the previous recent version
#
# after make package completes, copy all content of public/* into the github pages repository
# ----------------------------------
#
# User-Friendly Latest Release Downloads Links:
#
# cd latest/darwin-amd64/; chmod +x cp-remote; tar -czvf cp-remote.tar.gz cp-remote;
# cd ../../latest/linux-386/; chmod +x cp-remote; tar -czvf cp-remote.tar.gz cp-remote;
# cd ../../latest/linux-amd64/; chmod +x cp-remote; tar -czvf cp-remote.tar.gz cp-remote;
# cd ../../latest/windows-386/; zip -r cp-remote.zip cp-remote.exe;
# cd ../../latest/windows-amd64/; zip -r cp-remote.zip cp-remote.exe;/Users/azucca/dev/cp-remote-updater-gh-pages/
#
# commit and push the remote-environment-client gh-pages branch
#
#
# ----------------------------------
#
# Home Brew Tools:
#
# Update the homebrew-tools repo at https://github.com/continuouspipe/homebrew-tools/blob/master/Formula/cp-remote.rb#L6-L7
#
#


BINARY=cp-remote
VERSION=0.1.1-rc.2

GOROOT_FINAL=/usr/local/Cellar/go/1.7.4_2/libexec
CONFIG_PKG=github.com/continuouspipe/remote-environment-client/config
LDFLAGS=-ldflags="-X ${CONFIG_PKG}.CurrentVersion=${VERSION}"

#overwrite this to build for different arch and operative systems
BUILDOS=darwin
BUILDARCH=amd64

build:
	mkdir bin 2>/dev/null; true
	export GOROOT_FINAL=${GOROOT_FINAL}
	env GOOS=${BUILDOS} GOARCH=${BUILDARCH} go build ${LDFLAGS} -o bin/${BUILDOS}-${BUILDARCH}

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
