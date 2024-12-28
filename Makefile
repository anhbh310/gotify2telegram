BUILDDIR=./build
PLUGIN_NAME=telegram-plugin
PLUGIN_ENTRY=plugin.go
GO_VERSION=`cat $(BUILDDIR)/gotify-server-go-version`
DOCKER_BUILD_IMAGE=gotify/build
DOCKER_WORKDIR=/proj
DOCKER_RUN=docker run --rm -v "$$PWD/.:${DOCKER_WORKDIR}" -v "`go env GOPATH`/pkg/mod/.:/go/pkg/mod:ro" -w ${DOCKER_WORKDIR}
DOCKER_GO_BUILD=go build -mod=readonly -a -installsuffix cgo -ldflags "$$LD_FLAGS" -buildmode=plugin 

download-tools:
	go get -u github.com/gotify/plugin-api/cmd/gomod-cap

create-build-dir:
	mkdir -p ${BUILDDIR} || true

update-go-mod: create-build-dir
	GOTIFY_COMMIT=$(shell curl -s https://api.github.com/repos/gotify/server/git/ref/tags/v${GOTIFY_VERSION} | jq -r '.object.sha') && \
	wget -O ${BUILDDIR}/gotify-server.mod https://raw.githubusercontent.com/gotify/server/$${GOTIFY_COMMIT}/go.mod
	go run github.com/gotify/plugin-api/cmd/gomod-cap -from ${BUILDDIR}/gotify-server.mod -to go.mod
	rm ${BUILDDIR}/gotify-server.mod || true
	go mod tidy

get-gotify-server-go-version: create-build-dir
	rm -f ${BUILDDIR}/gotify-server-go-version || true
	GOTIFY_COMMIT=$(shell curl -s https://api.github.com/repos/gotify/server/git/ref/tags/v${GOTIFY_VERSION} | jq -r '.object.sha') && \
	wget -O ${BUILDDIR}/gotify-server-go-version https://raw.githubusercontent.com/gotify/server/$${GOTIFY_COMMIT}/GO_VERSION

build-linux-amd64: get-gotify-server-go-version update-go-mod
	${DOCKER_RUN} ${DOCKER_BUILD_IMAGE}:$(GO_VERSION)-linux-amd64 ${DOCKER_GO_BUILD} -o ${BUILDDIR}/${PLUGIN_NAME}-linux-amd64-v${GOTIFY_VERSION}${FILE_SUFFIX}.so ${DOCKER_WORKDIR}

build-linux-arm-7: get-gotify-server-go-version update-go-mod
	${DOCKER_RUN} ${DOCKER_BUILD_IMAGE}:$(GO_VERSION)-linux-arm-7 ${DOCKER_GO_BUILD} -o ${BUILDDIR}/${PLUGIN_NAME}-linux-arm-7-v${GOTIFY_VERSION}${FILE_SUFFIX}.so ${DOCKER_WORKDIR}

build-linux-arm64: get-gotify-server-go-version update-go-mod
	${DOCKER_RUN} ${DOCKER_BUILD_IMAGE}:$(GO_VERSION)-linux-arm64 ${DOCKER_GO_BUILD} -o ${BUILDDIR}/${PLUGIN_NAME}-linux-arm64-v${GOTIFY_VERSION}${FILE_SUFFIX}.so ${DOCKER_WORKDIR}

build: build-linux-arm-7 build-linux-amd64 build-linux-arm64

.PHONY: build
