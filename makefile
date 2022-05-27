all: build test

# This how we want to name the binary output
BINARY=httpgo

# These are the values we want to pass for VERSION and BUILD
# git tag 1.0.1
# git commit -am "One more change after the tags"
VERSION=`git describe --tags`
BUILD=`date +%FT%T%z`
BRANCH=`git branch --show-current`

# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS=-ldflags "-s -w -X main.Version=${VERSION} -X main.Build=${BUILD} -X main.Branch=${BRANCH}"

define increment
	$(eval v := $(shell git describe --tags --abbrev=0 | sed -Ee 's/^v|-.*//'))
    $(eval n := $(shell echo $(v) | awk -F. -v OFS=. -v f=$1 '{ $$f++ } 1'))
    @git tag -a v$(n) -m "Bumped to version $(n), $(m)"
	@git push
	@git push --tags
	@echo "Updating version $(v) to $(n)"
endef

update:
	$(call increment,3,path)
# Builds the project
run:
	go generate
	go run ${LDFLAGS} main.go -debug -port 8080
# Builds the project
build:
	go generate
	go build -i ${LDFLAGS} -o ${BINARY}
test:
	go test -v ./... > last_test.log
dto:
	go test ./apis
mod:
	go mod tidy
	sudo chown ruslan:progs go.*
# Builds the project
linux:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -i ${LDFLAGS} -o ${BINARY}

# Installs our project: copies binaries

install:
	go install ${LDFLAGS}

# Cleans our project: deletes binaries

clean:
	if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi
