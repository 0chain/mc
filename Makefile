MINIOPATH=$(GOPATH)/src/github.com/minio-io/mc

all: getdeps install

checkdeps:
	@echo "Checking deps.."
	@(env bash $(PWD)/buildscripts/checkdeps.sh)

checkgopath:
	@echo "Checking if project is at ${MINIOPATH}"
	@if [ ! -d ${MINIOPATH} ]; then echo "Project not found in $GOPATH, please follow instructions provided at https://github.com/Minio-io/minio/blob/master/CONTRIBUTING.md#setup-your-minio-github-repository" && exit 1; fi

getdeps: checkdeps checkgopath
	@go get github.com/tools/godep && echo "Installed godep"
	@go get github.com/golang/lint/golint && echo "Installed golint"
	@go get golang.org/x/tools/cmd/vet && echo "Installed vet"

verifiers: getdeps vet fmt lint

vet:
	@echo "Running $@"
	@go vet ./...
fmt:
	@echo "Running $@"
	@test -z "$$(gofmt -s -l . | grep -v Godeps/_workspace/src/ | tee /dev/stderr)" || \
		echo "+ please format Go code with 'gofmt -s'"
lint:
	@echo "Running $@"
	@test -z "$$(golint ./... | grep -v Godeps/_workspace/src/ | tee /dev/stderr)"

build-all: getdeps verifiers
	@echo "Building Libraries"
	@godep go generate ./...
	@godep go build ./...

test-all: build-all
	@echo "Running Test Suites:"
	@godep go test -race ./...

save: restore
	@godep save ./...

restore:
	@godep restore

env:
	@godep go env

docs-deploy:
	@mkdocs gh-deploy --clean

install: test-all
	@godep go install github.com/minio-io/mc && echo "Installed mc"
	@mkdir -p $(HOME)/.minio/mc
