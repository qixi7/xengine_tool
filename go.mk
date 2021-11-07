OS           := $(shell uname -s)
ifeq ($(OS), Linux)
APPBIN       := $(NAMEPREFIX))
APPBIN_SERVER:= $(NAMEPREFIX)_server
BUILDTIME    := $(shell date -Iseconds)
else ifeq ($(OS), Darwin)
APPBIN       := $(NAMEPREFIX))
APPBIN_SERVER:= $(NAMEPREFIX)_server
BUILDTIME    := $(shell date)
else
APPBIN       := $(NAMEPREFIX).exe
APPBIN_SERVER:= $(NAMEPREFIX)_server.exe
BUILDTIME    := $(shell date -Iseconds)
endif

ifeq ($(shell [ -f .svnversion ] && echo OK), OK)
CODEVER       := $(shell . ./.svnversion; echo "$$CODEVER")
else
CODEVER       := $(shell git log | grep "^commit" | head -1)
endif

OUTPUT		 := $(CURDIR)/output
SERVERDATA   := $(OUTPUT)/serverdata
LDFLAGS      := -ldflags "-X 'xcore/xutil.GitVersion=$(CODEVER)' -X 'xcore/xutil.BuildTime=$(BUILDTIME)' -X xcore/xutil.BuildType=Testing"
LDFLAGS_PROD := -ldflags "-X 'xcore/xutil.GitVersion=$(CODEVER)' -X 'xcore/xutil.BuildTime=$(BUILDTIME)' -X xcore/xutil.BuildType=Production"
REBUILDFLAGS := -a
GCFLAGS      := -gcflags "all=-N -l"
PACKAGE      := -i

GOOS	:=linux
GOARCH	:=amd64
CGO_ENABLED:= 0

.PHONY: help
help:
	$(info Available targets: )
	$(info help             )
	$(info build      build debug version)
	$(info rebuild    rebuild debug version)
	$(info release    build release version)
	$(info clean            )
	$(info test             )
	$(info install    build release version to output)

.PHONY: build
build:
	GOPATH=$(GOPATH) CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(APPBIN_SERVER) $(GCFLAG) $(LDFLAGS) $(SRC)

.PHONY: release
release:
	GOPATH=$(GOPATH) CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(APPBIN) $(LDFLAGS) $(REBUILDFLAGS) $(SRC)

.PHONY: rebuild
rebuild:
	GOPATH=$(GOPATH) CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(APPBIN) $(GCFLAG) $(LDFLAGS) $(REBUILDFLAGS) $(SRC)

.PHONY: test
test:
	GOPATH=$(GOPATH) go test $(SRC)/...

.PHONY: install
install: $(OUTPUT)/$(APPBIN_SERVER)

$(OUTPUT): mkdir $(OUTPUT)

$(OUTPUT)/$(APPBIN_SERVER): $(OUTPUT)
	GOPATH=$(GOPATH) go build -o $(OUTPUT)/$(APPBIN_SERVER) $(LDFLAGS_PROD) $(REBUILDFLAGS) $(SRC)

.PHONY: data
data: $(SERVERDATA)

$(SERVERDATA): $(OUTPUT)
	mkdir -p $(SERVERDATA)
	[ -d ../serverdata ] && rsync -ax --delete ../serverdata $(OUTPUT)

.PHONY: clean
clean:
	rm -rf $(OUTPUT)
	GOPATH=$(GOPATH) go clean ./...
