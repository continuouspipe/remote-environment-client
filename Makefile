# In order release a new version, checkout the content of the bucket inviqa-cp-remote-client-environment
# all the GOOS-GOARCH.json files e.g.(darwin-amd64.json, windows-386.json, and so on..) along with the previous
# 2 releases folders e.g.(0.0.1, 0.0.2 if you are releasing 0.0.3) and copy this in a folder called "public"
#
# Then cross-compile for all supported goos-goarch:
#
# $ make build BUILDOS=darwin BUILDARCH=amd64
# $ make build BUILDOS=windows BUILDARCH=amd64
# $ make build BUILDOS=windows BUILDARCH=386
# $ make build BUILDOS=linux BUILDARCH=amd64
# $ make build BUILDOS=linux BUILDARCH=386
#
#
# Auto-Upgrade Section:
#
# make build will put the binaries into the bin/ folder.
# After this is done run "make package" which will put the new binaries into public/ along with the binary diffs
# in order to have a quicker upgrade from the previous recent version
#
# after make package completes, copy and sync all content of public/* into the aws s3 bucket inviqa-cp-remote-client-environment "downloads" folder
#
# $ aws s3 sync downloads/x.y.x (-n)/ s3://inviqa-cp-remote-client-environment/downloads/x.y.x (-n)*
# $ aws s3 sync downloads/x.y.x (-1)/ s3://inviqa-cp-remote-client-environment/downloads/x.y.x (-1)*
# $ aws s3 sync downloads/x.y.x/ s3://inviqa-cp-remote-client-environment/downloads/x.y.x
# $ aws s3 sync downloads s3://inviqa-cp-remote-client-environment/downloads/ --exclude=* --include=*.json
#
# as (-1) and (-n) means any previous version for which there is a bin diff that needs uploaded
#
# ----------------------------------
#
# User-Friendly Latest Release Downloads Links:
#
# $ cp downloads/0.1.4/darwin-amd64.gz latest/darwin-amd64/;
# $ cp downloads/0.1.4/linux-amd64.gz latest/linux-amd64/;
# $ cp downloads/0.1.4/linux-386.gz latest/linux-386/;
# $ cp downloads/0.1.4/windows-amd64.gz latest/windows-amd64/;
# $ cp downloads/0.1.4/windows-386.gz latest/windows-386/;
#
# $ cd latest/darwin-amd64/; gzip -d darwin-amd64.gz; chmod +x darwin-amd64; mv darwin-amd64 cp-remote;
# $ tar -czvf cp-remote.tar.gz cp-remote;
#
# $ cd ../../latest/linux-amd64/; gzip -d linux-amd64.gz; chmod +x linux-amd64; mv linux-amd64 cp-remote;
# $ tar -czvf cp-remote.tar.gz cp-remote;
#
# $ cd ../../latest/linux-386/; gzip -d linux-386.gz; chmod +x linux-386; mv linux-386 cp-remote;
# $ tar -czvf cp-remote.tar.gz cp-remote;
#
# $ cd ../../latest/windows-386/; gzip -d windows-386.gz; mv windows-386 cp-remote.exe;
# $ zip -r cp-remote.zip cp-remote.exe;
#
# $ cd ../../latest/windows-amd64/; gzip -d windows-amd64.gz; mv windows-amd64 cp-remote.exe;
# $ zip -r cp-remote.zip cp-remote.exe;
#
# $ cd ../../
# $ rm -fr latest/*/cp-remote
# $ rm -fr latest/*/cp-remote.exe
#
# $ ls -l latest/*/*
#
# sync the new files into the aws s3 bucket inviqa-cp-remote-client-environment
# $ aws s3 sync latest s3://inviqa-cp-remote-client-environment/latest
# ----------------------------------
#
# Home Brew Tools:
#
# Update the homebrew-tools repo at https://github.com/continuouspipe/homebrew-tools/blob/master/Formula/cp-remote.rb#L6-L7
#
#


BINARY=cp-remote
VERSION=0.1.5

CONFIG_PKG=github.com/continuouspipe/remote-environment-client/config
LDFLAGS=-ldflags="-X ${CONFIG_PKG}.CurrentVersion=${VERSION}"
TRIMFLAGS=-gcflags=-trimpath=${GOPATH} -asmflags=-trimpath=${GOPATH}

#overwrite this to build for different arch and operative systems
BUILDOS=darwin
BUILDARCH=amd64

build:
	mkdir bin 2>/dev/null; true
	env GOOS=${BUILDOS} GOARCH=${BUILDARCH} go build ${LDFLAGS} ${TRIMFLAGS} -o bin/${BUILDOS}-${BUILDARCH}

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

# copy from the aws s3 bucket inviqa-cp-remote-client-environment the latest version and json file into a public/ folder before running this command
# as go-selfupdate needs to find the previous installation in order to generate the binary diff for incremental upgrades
package:
	go-selfupdate bin/ ${VERSION}
