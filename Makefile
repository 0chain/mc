## filter multiple GOPATH
all: getdeps install

checkdeps:
	@echo "Checking deps:"
	@(env bash $(PWD)/buildscripts/checkdeps.sh)

checkgopath:
	@echo "Checking if project is at ${GOPATH}"
	@for mcpath in $(echo ${GOPATH} | sed 's/:/\n/g'); do if [ ! -d ${mcpath}/src/github.com/minio/mc ]; then echo "Project not found in ${mcpath}, please follow instructions provided at https://github.com/minio/minio/blob/master/CONTRIBUTING.md#setup-your-minio-github-repository" && exit 1; fi done

getdeps: checkdeps checkgopath
	@go get github.com/golang/lint/golint && echo "Installed golint:"
	@go get golang.org/x/tools/cmd/vet && echo "Installed vet:"
	@go get github.com/fzipp/gocyclo && echo "Installed gocyclo:"
	@go get github.com/remyoudompheng/go-misc/deadcode && echo "Installed deadcode:"

# verifiers: getdeps vet fmt lint cyclo deadcode
verifiers: getdeps vet fmt lint cyclo deadcode

vet:
	@echo "Running $@:"
	@go vet .
	@go vet github.com/minio/mc/pkg...
fmt:
	@echo "Running $@:"
	@gofmt -s -l *.go
	@gofmt -s -l pkg
lint:
	@echo "Running $@:"
	@golint .
	@golint pkg

cyclo:
	@echo "Running $@:"
	@gocyclo -over 30 .

deadcode:
	@echo "Running $@:"
	@deadcode

build: getdeps verifiers
	@echo "Installing mc:"
	@go test -race ./
	@go test -race github.com/minio/mc/pkg...

gomake-all: build
	@go install github.com/minio/mc
	@mkdir -p $(HOME)/.mc

release: genversion
	@echo "Installing mc with new version.go:"
	@go install github.com/minio/mc
	@mkdir -p $(HOME)/.mc

genversion:
	@echo "Generating a new version.go:"
	@go run genversion.go

coverage:
	@go test -race -coverprofile=cover.out
	@go tool cover -html=cover.out && echo "Visit your browser"

install: gomake-all

clean:
	@rm -fv cover.out
	@rm -fv mc
