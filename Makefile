VERSION := $(shell echo $(shell git describe --tags) | sed 's/^v//')
# SDKVERSION := $(shell go list -m -u -f '{{.Version}}' github.com/cosmos/cosmos-sdk)
TMVERSION := $(shell go list -m -u -f '{{.Version}}' github.com/tendermint/tendermint)
COMMIT  := $(shell git log -1 --format='%H')

TESTPACKAGES := $(shell go list ./... | grep -v testnet) 

all: install

LD_FLAGS = -X github.com/dauTT/valink/cmd.Version=$(VERSION) \
	-X github.com/dauTT/valink/cmd.Commit=$(COMMIT) \
	-X github.com/dauTT/valink/cmd.SDKVersion=$(SDKVERSION) \
	-X github.com/dauTT/valink/cmd.TMVersion=$(TMVERSION)

BUILD_FLAGS := -ldflags '$(LD_FLAGS)'

# https://grpc.io/docs/protoc-installation/
# https://github.com/protocolbuffers/protobuf/releases/download/v3.11.2/protoc-3.11.2-linux-x86_64.zip
proto-gen:
	@echo "Generating Protobuf files:"
	protoc --go_out=plugins=grpc:./  proto/*.proto

# run "make proto-gen" before using the build rule
build:
	@echo "Build valink binary:"
	go build -mod readonly $(BUILD_FLAGS) -o build/valink ./valink

build-all: proto-gen build
	
install:
	go install -mod readonly $(BUILD_FLAGS) ./valink

build-linux:
	GOOS=linux GOARCH=amd64 go build --mod readonly $(BUILD_FLAGS) -o ./build/valink ./valink

test:
	go test -mod=readonly $(TESTPACKAGES) -count=1

race:
	go test -race -short -mod=readonly $(TESTPACKAGES) -count=1

tools:
	go install golang.org/x/lint/golint

clean:
	rm -rf build

build-valink-docker: proto-gen
	docker rmi dautt/valink:vgRPC2.0.1
	docker build . -t dautt/valink:vgRPC2.0.1

.PHONY: all lint test race msan tools clean build proto-gen