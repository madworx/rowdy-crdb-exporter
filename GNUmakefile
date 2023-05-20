GO := CGO_ENABLED=0 go
DIST_DIR := dist
TARGETS := rowdy_linux_amd64_exporter rowdy_windows_amd64_exporter.exe rowdy_linux_arm64_exporter
TARGETS_DIST := $(addprefix $(DIST_DIR)/, $(TARGETS))
BUILD_CMD := $(GO) build -ldflags="-X main.gitCommit=$(shell git rev-parse HEAD) -X main.gitTag=$(shell git describe --tags --abbrev=0 2>/dev/null)" -o

.PHONY: all clean build-all test

all: build-all

$(DIST_DIR)/rowdy_linux_amd64_exporter: main.go .gomodtidy | $(DIST_DIR)
	GOARCH=amd64 GOOS=linux $(BUILD_CMD) $@ 

$(DIST_DIR)/rowdy_windows_amd64_exporter.exe: main.go .gomodtidy | $(DIST_DIR)
	GOARCH=amd64 GOOS=windows $(BUILD_CMD) $@ 

$(DIST_DIR)/rowdy_linux_arm64_exporter: main.go .gomodtidy | $(DIST_DIR)
	GOARCH=arm64 GOOS=linux $(BUILD_CMD) $@

$(DIST_DIR):
	mkdir -p $(DIST_DIR)

$(DIST_DIR)/SHA256SUM.txt: $(TARGETS_DIST) | $(DIST_DIR)
	cd $(DIST_DIR) && sha256sum $(TARGETS) > SHA256SUM.txt

build-all: $(TARGETS_DIST) $(DIST_DIR)/SHA256SUM.txt

test:
	go test -v ./... -coverprofile=coverage.out -covermode=atomic
	go tool cover -html=coverage.out -o coverage.html

e2e_tests/app: GNUmakefile main.go main_test.go .gomodtidy | $(DIST_DIR)
	GOARCH=amd64 GOOS=linux $(GO) build -cover -o e2e_tests/app 

.PHONY: coverage-e2e.out
coverage-e2e.out:
	docker compose exec -u $(shell id -u) rowdy make e2e_tests/app
	docker compose run --rm robotframework .
	docker compose exec -u $(shell id -u) rowdy go tool covdata textfmt -i=./e2e_tests -o=coverage-e2e.out
	docker compose exec -u $(shell id -u) rowdy go tool cover -html=coverage-e2e.out -o coverage-e2e.html
	docker compose exec -u $(shell id -u) rowdy go tool covdata percent -i=./e2e_tests

clean:
	rm -rf $(DIST_DIR) e2e_tests/covcounters.* e2e_tests/covmeta.* coverage-e2e.out coverage-e2e.html coverage.out coverage.html

go.mod:
	$(GO) mod init github.com/madworx/rowdy-crdb-exporter

go.sum: go.mod
	$(GO) mod tidy
	touch go.sum

.gomodtidy: go.sum
	@touch .gomodtidy
