GO := CGO_ENABLED=0 go
DIST_DIR := dist
TARGETS := rowdy_linux_amd64_exporter rowdy_windows_amd64_exporter.exe rowdy_linux_arm64_exporter
TARGETS_DIST := $(addprefix $(DIST_DIR)/, $(TARGETS))
BUILD_CMD := $(GO) build -ldflags="-X main.gitCommit=$(shell git rev-parse HEAD) -X main.gitTag=$(shell git describe --tags --abbrev=0)" -o

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

clean:
	rm -rf $(DIST_DIR)

go.mod:
	$(GO) mod init github.com/madworx/rowdy-crdb-exporter

go.sum: go.mod
	$(GO) mod tidy
	touch go.sum

.gomodtidy: go.sum
	@touch .gomodtidy
