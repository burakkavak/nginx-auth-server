PROJECT_NAME			:= nginx-auth-server

LINUX_386_BINARY		:= ${PROJECT_NAME}-linux-i386
LINUX_AMD64_BINARY		:= ${PROJECT_NAME}-linux-amd64
LINUX_ARM_BINARY		:= ${PROJECT_NAME}-linux-arm
LINUX_ARM64_BINARY		:= ${PROJECT_NAME}-linux-arm64
WINDOWS_AMD64_BINARY	:= ${PROJECT_NAME}.exe

ALL_BINARIES			:= ${LINUX_386_BINARY} ${LINUX_AMD64_BINARY} ${LINUX_ARM_BINARY} ${LINUX_ARM64_BINARY} ${WINDOWS_AMD64_BINARY}

REQUIRED_BINS			:= go tar npm

$(foreach bin,$(REQUIRED_BINS),\
    $(if $(shell command -v $(bin) 2> /dev/null),$(),$(error Error: install '$(bin)')))

GOOS					:= $(shell go env GOOS)
GOARCH					:= $(shell go env GOARCH)

compile:
	npm i
	npm run build

	go build -o "./bin/${PROJECT_NAME}-${GOOS}-${GOARCH}" -tags "prod netgo" ./src

compileAll:
	npm i
	npm run build

	GOOS=linux GOARCH=386 go build -o ./bin/${LINUX_386_BINARY} -tags "prod netgo" ./src
	GOOS=linux GOARCH=amd64 go build -o ./bin/${LINUX_AMD64_BINARY} -tags "prod netgo" ./src
	GOOS=linux GOARCH=arm go build -o ./bin/${LINUX_ARM_BINARY} -tags "prod netgo" ./src
	GOOS=linux GOARCH=arm64 go build -o ./bin/${LINUX_ARM64_BINARY} -tags "prod netgo" ./src
	GOOS=windows GOARCH=amd64 go build -o ./bin/${WINDOWS_AMD64_BINARY} -tags "prod netgo" ./src

package:
	cd bin && $(foreach binary,$(ALL_BINARIES),tar cfz $(binary).tar.gz $(binary);)
