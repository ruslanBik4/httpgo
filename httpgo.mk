# These are the values we want to pass for VERSION and BUILD
# git tag 1.0.1
# git commit -am "One more change after the tags"
GO_VERSION=`go version | sed -E 's/^go version go([0-9.]+).*/\1/'`
OS_VERSION := $(shell sh -c 'uname -srm | awk '\''{print $$1 "-" $$2}'\''')
HTTPGO_VER=$(shell git -C "$(go list -m -f '{{.Dir}}' github.com/ruslanBik4/httpgo)" describe --tags --abbrev=0)
#get version, branch &build time of app
VERSION=`git describe --tags --abbrev=0`
BUILD=`date +%FT%T%z`
BRANCH=`git branch --show-current`
HTTPGO=github.com/ruslanBik4/httpgo/httpGo

HTTPGO_LDFLAGS = -X ${HTTPGO}.HTTPGOVer=${HTTPGO_VER} -X ${HTTPGO}.OSVersion=${OS_VERSION} -X ${HTTPGO}.GoVersion=${GO_VERSION} -X ${HTTPGO}.Version=${VERSION} -X ${HTTPGO}.Build=${BUILD} -X ${HTTPGO}.Branch=${BRANCH}
