all: getdeps install

checkdeps:
	@./checkdeps.sh

createsymlink:
	@mkdir -p $(GOPATH)/src/github.com/minio-io/;
	@if test ! -e $(GOPATH)/src/github.com/minio-io/mc; then echo "Creating symlink to $(GOPATH)/src/github.com/minio-io/mc" && ln -s $(PWD) $(GOPATH)/src/github.com/minio-io/mc; fi

getdeps: checkdeps
	@go get github.com/tools/godep && echo "Installed godep"
	@go get golang.org/x/tools/cmd/cover && echo "Installed cover"

build-all: getdeps createsymlink
	@echo "Building Libraries"
	@godep go generate ./...
	@godep go build ./...

test-all: build-all
	@echo "Running Test Suites:"
	@godep go test -race ./...

install: test-all
	@godep go install github.com/minio-io/mc && echo "Installed mc"
