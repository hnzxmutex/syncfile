GOPATH := $(shell pwd)
GOBIN := $(GOPATH)/bin
PROJECT_NAME := syncfile
.PHONY: clean

all:
	@GOPATH=$(GOPATH) && echo "install at $(GOBIN)" && cd src/$(PROJECT_NAME) && go get && cd $(GOPATH) && go install $(PROJECT_NAME)

linux:
	@GOPATH=$(GOPATH) && echo "build at $(GOBIN)" && cd src/$(PROJECT_NAME) && go get && cd $(GOPATH) && GOOS=linux GOARCH=amd64 go build  -o ./bin/$(PROJECT_NAME).linux $(PROJECT_NAME)

windows:
	@GOPATH=$(GOPATH) && echo "build at $(GOBIN)" && cd src/$(PROJECT_NAME) && go get && cd $(GOPATH) && GOOS=windows GOARCH=amd64 go build  -o ./bin/$(PROJECT_NAME).exe $(PROJECT_NAME)

test:
	@GOPATH=$(GOPATH) && echo "testing" && cd src/$(PROJECT_NAME) && go get && cd $(GOPATH) && go test $(PROJECT_NAME)

clean:
	@rm -fr bin pkg *.log *.pid
