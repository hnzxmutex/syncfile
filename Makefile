GOPATH := $(shell pwd)
GOBIN := $(GOPATH)/bin
PROJECT_NAME := syncfile
BUILD_DATE := $(shell date +"%Y.%m.%d-%H:%M:%S")
BUILD_FLAGS := -ldflags "-X syncfile/cmd.BuildDate="$(BUILD_DATE)
.PHONY: clean

mac:
	@GOPATH=$(GOPATH) && echo "build at $(GOBIN)" && cd src/$(PROJECT_NAME) && go get && cd $(GOPATH) && go build $(BUILD_FLAGS) -o ./bin/$(PROJECT_NAME) $(PROJECT_NAME)

all: mac linux windows

linux:
	@GOPATH=$(GOPATH) && echo "build linux.app at $(GOBIN)" && cd src/$(PROJECT_NAME) &&GOOS=linux GOARCH=amd64 go get && cd $(GOPATH) && GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o ./bin/$(PROJECT_NAME).linux $(PROJECT_NAME)

windows:
	@GOPATH=$(GOPATH) && echo "build windows.app at $(GOBIN)" && cd src/$(PROJECT_NAME) &&GOOS=windows go get && cd $(GOPATH) && GOOS=windows GOARCH=amd64 go build $(BUILD_FLAGS) -o ./bin/$(PROJECT_NAME).exe $(PROJECT_NAME)

test:
	@GOPATH=$(GOPATH) && echo "testing" && cd src/$(PROJECT_NAME) && go get && cd $(GOPATH) && go test $(PROJECT_NAME)

clean:
	@rm -fr pkg *.log *.pid
