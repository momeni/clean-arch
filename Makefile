
TARGET := caweb

.PHONY: all build integration-tests
all: build

build:
	go build -tags="go_json nomsgpack" ./cmd/caweb

integration-tests:
	if [ ! -s $$XDG_RUNTIME_DIR/podman/podman.sock ]; then \
		systemctl --user start podman.service; \
	fi && \
	DOCKER_HOST=unix://$$XDG_RUNTIME_DIR/podman/podman.sock \
	go test ./pkg/adapter/restful/gin -run TestIntegration

.PHONY: install-staticcheck install-revive revive lint
install-staticcheck:
	go install honnef.co/go/tools/cmd/staticcheck@2023.1.6

install-revive:
	go get -u github.com/mgechev/revive

revive:
	revive -formatter friendly ./...

lint:
	@staticcheck ./...
	@revive ./...
