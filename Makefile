VERSION := $(shell echo $(shell git describe --tags) | sed 's/^v//')
# SDKVERSION := $(shell go list -m -u -f '{{.Version}}' github.com/cosmos/cosmos-sdk)
TMVERSION := $(shell go list -m -u -f '{{.Version}}' github.com/tendermint/tendermint)
COMMIT  := $(shell git log -1 --format='%H')

TESTPACKAGES := $(shell go list ./... | grep -v testnet) 

all: install

LD_FLAGS = -X github.com/dauTT/tendermint-mpc-validator/cmd.Version=$(VERSION) \
	-X github.com/dauTT/tendermint-mpc-validator/cmd.Commit=$(COMMIT) \
	-X github.com/dauTT/tendermint-mpc-validator/cmd.SDKVersion=$(SDKVERSION) \
	-X github.com/dauTT/tendermint-mpc-validator/cmd.TMVersion=$(TMVERSION)

BUILD_FLAGS := -ldflags '$(LD_FLAGS)'

build:
	go build -mod readonly $(BUILD_FLAGS) -o build/valink ./valink

install:
	go install -mod readonly $(BUILD_FLAGS) ./valink

build-linux:
	GOOS=linux GOARCH=amd64 go build --mod readonly $(BUILD_FLAGS) -o ./build/valink ./valink

test:
	go test -mod=readonly $(TESTPACKAGES) -count=1

race:
	go test -race -short -mod=readonly $(TESTPACKAGES) -count=1

#msan:
#	go test -msan -short -mod=readonly $(TESTPACKAGES) -count=1

tools:
	go install golang.org/x/lint/golint

clean:
	rm -rf build

# build-valink-docker:
# 	docker build -t dauTT/valink:$(VERSION) -f ./Dockerfile .

# push-junod-docker:
# 	docker push dauTT/junod:$(SDKVERSION)




.PHONY: all lint test race msan tools clean build