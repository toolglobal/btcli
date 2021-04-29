ifeq ($(mode),debug)
	LDFLAGS="-X 'main.BUILD_TIME=`date`' -X 'main.GO_VERSION=`go version`' -X main.GIT_HASH=`git rev-parse HEAD`"
else
	LDFLAGS="-s -w -X 'main.BUILD_TIME=`date`' -X 'main.GO_VERSION=`go version`' -X main.GIT_HASH=`git rev-parse HEAD`"
endif

.PHONY: build
build:
	export GOPROXY="https://goproxy.io,direct"
	go build -ldflags ${LDFLAGS} -o build/btcli *.go
	cp config.toml build

clean:
	rm -rf ./build

.PHONY: docs
docs:
	swag init -g main.go
