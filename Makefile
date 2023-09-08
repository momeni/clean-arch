
TARGET := caweb

all: build

build:
	go build -tags="go_json nomsgpack" ./cmd/caweb

.PHONY: all build
