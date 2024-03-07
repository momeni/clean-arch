
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

.PHONY: config-test
config-test:
	go test ./pkg/adapter/config/...

.PHONY: mig-tests
mig-tests:
	if [ ! -s $$XDG_RUNTIME_DIR/podman/podman.sock ]; then \
		systemctl --user start podman.service; \
	fi && \
	DOCKER_HOST=unix://$$XDG_RUNTIME_DIR/podman/podman.sock \
	go test ./pkg/core/usecase/migrationuc

.PHONY: scram-test
scram-test:
	go test ./pkg/adapter/hash/scram

.PHONY: test
test: config-test integration-tests mig-tests scram-test

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

SRC_DB_DIR := dist/.db/caweb1_0_0
.PHONY: src-db src-db-psql
src-db: $(SRC_DB_DIR)/.pgpass
	podman start caweb1_0_0-pg16-dbms

$(SRC_DB_DIR)/.pgpass:
	adminpass="$$(head -c16 /dev/random | sha1sum | cut -d' ' -f1)" && \
		cawebpass="$$(head -c16 /dev/random | sha1sum | cut -d' ' -f1)" && \
		mkdir -p $(SRC_DB_DIR)/data && \
		echo "127.0.0.1:5455:caweb1_0_0:admin:$$adminpass" > $@ && \
		echo "127.0.0.1:5455:caweb1_0_0:caweb:$$cawebpass" >> $@ && \
		chmod 0600 $@ && \
		podman run -t --detach --replace --name caweb1_0_0-pg16-dbms \
			-e POSTGRES_USER="admin" \
			-e POSTGRES_PASSWORD="$$adminpass" \
			-e POSTGRES_DB="caweb1_0_0" \
			-e POSTGRES_HOST_AUTH_METHOD="scram-sha-256" \
			-e POSTGRES_INITDB_ARGS="--auth-host=scram-sha-256" \
			-v $(CURDIR)/$(SRC_DB_DIR)/data:/var/lib/postgresql/data:Z \
			-p 5455:5432 \
			docker.io/postgres:16-bookworm

src-db-psql: src-db
	PGPASSFILE=$(SRC_DB_DIR)/.pgpass \
		psql -h 127.0.0.1 -p 5455 -U admin -d caweb1_0_0

DST_DB_DIR := dist/.db/caweb1_1_0
.PHONY: dst-db dst-db-psql
dst-db: $(DST_DB_DIR)/.pgpass
	podman start caweb1_1_0-pg16-dbms

$(DST_DB_DIR)/.pgpass:
	adminpass="$$(head -c16 /dev/random | sha1sum | cut -d' ' -f1)" && \
		cawebpass="$$(head -c16 /dev/random | sha1sum | cut -d' ' -f1)" && \
		mkdir -p $(DST_DB_DIR)/data && \
		echo "127.0.0.1:5456:caweb1_1_0:admin:$$adminpass" > $@ && \
		echo "127.0.0.1:5456:caweb1_1_0:caweb:$$cawebpass" >> $@ && \
		chmod 0600 $@ && \
		podman run -t --detach --replace --name caweb1_1_0-pg16-dbms \
			-e POSTGRES_USER="admin" \
			-e POSTGRES_PASSWORD="$$adminpass" \
			-e POSTGRES_DB="caweb1_1_0" \
			-e POSTGRES_HOST_AUTH_METHOD="scram-sha-256" \
			-e POSTGRES_INITDB_ARGS="--auth-host=scram-sha-256" \
			-v $(CURDIR)/$(DST_DB_DIR)/data:/var/lib/postgresql/data:Z \
			-p 5456:5432 \
			docker.io/postgres:16-bookworm

dst-db-psql: dst-db
	PGPASSFILE=$(DST_DB_DIR)/.pgpass \
		psql -h 127.0.0.1 -p 5456 -U admin -d caweb1_1_0

.PHONY: grep
grep:
	grep -R --exclude-dir=.git --exclude-dir=dist ${ARGS} .

.PHONY: manual-migration-test
manual-migration-test: build
	podman stop caweb1_0_0-pg16-dbms
	podman stop caweb1_1_0-pg16-dbms
	for dir in "$(SRC_DB_DIR)" "$(DST_DB_DIR)"; do \
		podman unshare rm -rf "$$dir" && mkdir -p "$$dir"; \
	done
	$(MAKE) src-db dst-db
	sleep 5
	./$(TARGET) db init-dev --config configs/sample-src-config.yaml
	./$(TARGET) db migrate configs/sample-src-config.yaml \
		configs/sample-dst-config.yaml \
		--config configs/sample-config.yaml
	echo "Check configs/sample-config.yaml as the target config file."
