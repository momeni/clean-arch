
TARGET := caweb

all: build

build:
	go build -tags="go_json nomsgpack" ./cmd/caweb

integration-tests:
	if [ ! -s $$XDG_RUNTIME_DIR/podman/podman.sock ]; then \
		systemctl --user start podman.service; \
	fi && \
	DOCKER_HOST=unix://$$XDG_RUNTIME_DIR/podman/podman.sock \
	go test ./pkg/adapter/restful/gin -run TestIntegration

.PHONY: all build integration-tests
