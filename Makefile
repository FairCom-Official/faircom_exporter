.PHONY: build run test clean install lint fmt vet check deps help package-all package-tar package-rpm package-deb

VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "1.0.0")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT_SHA=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_ARCH=$(shell uname -m)
BUILD_ARCH_ALT=$(shell [ $(BUILD_ARCH) = "x86_64" ] && echo "amd64" || ([ "$(BUILD_ARCH)" = "aarch64" ] && echo "arm64" || echo "unknown"))
BINARY_BASENAME=faircom_exporter
BINARY_NAME=$(BINARY_BASENAME)
LDFLAGS=-ldflags "-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.CommitSHA=$(COMMIT_SHA)"
#DEBUGFLAGS=-gcflags=all="-N -l"
OUTPUT_DIR=output
FAIRCOMDB_DIR=/opt/faircom/drivers/ctree.drivers
OPENSSL_DIR=${FAIRCOMDB_DIR}/lib/License.Lib/openssl
PACKAGE_NAME_BASE=$(BINARY_BASENAME)-$(VERSION)
PACKAGE_NAME=$(PACKAGE_NAME_BASE)-linux-$(BUILD_ARCH)

## help: Display this help message
help:
	@echo "Available targets:"
	@awk '/^##/ {printf "\033[36m%-20s\033[0m %s\n", $$2, substr($$0, index($$0, $$3))}' $(MAKEFILE_LIST)

## build: Build the exporter binary for current platform
build: build-snapshot
	mkdir -p $(OUTPUT_DIR)
	CGO_ENABLED=1 go build $(LDFLAGS) -o $(OUTPUT_DIR)/$(BINARY_NAME) ./cmd/exporter


pkg/collector/snapshot.o: pkg/collector/c/snapshot.c
	gcc -g -c -fPIC -I${FAIRCOMDB_DIR}/include/unix/multithreaded/dynamic -I${FAIRCOMDB_DIR}/include -I${OPENSSL_DIR}/include -opkg/collector/snapshot.o pkg/collector/c/snapshot.c

pkg/collector/libsnapshot.a: pkg/collector/snapshot.o
	ar rcs pkg/collector/libsnapshot.a pkg/collector/snapshot.o

build-snapshot: pkg/collector/libsnapshot.a
	

## build-linux-amd64: Build for Linux AMD64
build-linux-amd64: build-snapshot
	mkdir -p $(OUTPUT_DIR)
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build $(DEBUGFLAGS) $(LDFLAGS) -o $(OUTPUT_DIR)/$(BINARY_NAME) ./cmd/exporter

## build-linux-arm64: Build for Linux ARM64
build-linux-arm64: build-snapshot
	mkdir -p $(OUTPUT_DIR)
	CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build $(DEBUGFLAGS) $(LDFLAGS) -o $(OUTPUT_DIR)/$(BINARY_NAME) ./cmd/exporter

## package-tar: Create tar.gz packages for Linux
package-tar: build
	mkdir -p $(OUTPUT_DIR)/tar
	mkdir -p $(OUTPUT_DIR)/tar/tmp
	mkdir -p $(OUTPUT_DIR)/tar/tmp/$(PACKAGE_NAME)
	cp $(OUTPUT_DIR)/$(BINARY_NAME) $(OUTPUT_DIR)/tar/tmp/$(PACKAGE_NAME)/$(BINARY_NAME)
	cp config.yaml $(OUTPUT_DIR)/tar/tmp/$(PACKAGE_NAME)/
	cp README.md $(OUTPUT_DIR)/tar/tmp/$(PACKAGE_NAME)/
	tar -czf $(OUTPUT_DIR)/tar/$(PACKAGE_NAME).tar.gz -C $(OUTPUT_DIR)/tar/tmp $(PACKAGE_NAME)
	rm -rf $(OUTPUT_DIR)/tar/tmp
	@echo "Tar packages created in $(OUTPUT_DIR)/tar/"

## package-deb: Create DEB packages for Debian/Ubuntu
package-deb: build
	@which fpm > /dev/null || (echo "ERROR: fpm not found. Install with: gem install fpm" && exit 1)
	mkdir -p $(OUTPUT_DIR)/deb
	fpm -s dir -t deb -n $(BINARY_NAME) -v $(VERSION) \
		-a $(BUILD_ARCH_ALT) \
		--description "Prometheus exporter for FairCom database metrics" \
		--url "https://github.com/faircom/prometheus-exporter" \
		--license "MIT" \
		--maintainer "FairCom" \
		--deb-systemd systemd/$(BINARY_NAME).service \
		--config-files /etc/$(BINARY_NAME)/config.yaml \
		--directories /etc/$(BINARY_NAME) \
		--package $(OUTPUT_DIR)/deb/$(PACKAGE_NAME).deb \
		$(OUTPUT_DIR)/$(BINARY_NAME)=/usr/local/bin/$(BINARY_NAME) \
		config.yaml=/etc/$(BINARY_NAME)/config.yaml
	@echo "DEB package created in $(OUTPUT_DIR)/deb/"

## package-rpm: Create RPM package for RHEL/CentOS
package-rpm: build
	@which fpm > /dev/null || (echo "ERROR: fpm not found. Install with: gem install fpm" && exit 1)
	mkdir -p $(OUTPUT_DIR)/rpm
	fpm -s dir -t rpm -n $(BINARY_NAME) -v $(VERSION) \
		-a $(BUILD_ARCH) \
		--description "Prometheus exporter for FairCom database metrics" \
		--url "https://github.com/faircom/prometheus-exporter" \
		--license "MIT" \
		--maintainer "FairCom" \
		--config-files /etc/$(BINARY_NAME)/config.yaml \
		--directories /etc/$(BINARY_NAME) \
		--package $(OUTPUT_DIR)/rpm/$(PACKAGE_NAME).rpm \
		$(OUTPUT_DIR)/$(BINARY_NAME)=/usr/local/bin/$(BINARY_NAME) \
		config.yaml=/etc/$(BINARY_NAME)/config.yaml \
		systemd/$(BINARY_NAME).service=/usr/lib/systemd/system/$(BINARY_NAME).service
	@echo "RPM package created in $(OUTPUT_DIR)/rpm/"

## package-all: Create all packages (tar.gz, deb, rpm)
package-all: package-tar package-deb package-rpm
	@echo ""
	@echo "All packages created:"
	@ls -lh $(OUTPUT_DIR)/tar/
	@ls -lh $(OUTPUT_DIR)/deb/
	@ls -lh $(OUTPUT_DIR)/rpm/

## run: Build and run the exporter
run: build
	$(OUTPUT_DIR)/$(BINARY_NAME) --config config.yaml

## test: Run tests
test:
	go test -v -race ./...

## test-coverage: Run tests with coverage
test-coverage:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

## clean: Clean build artifacts
clean:
	go clean
	rm -rf $(OUTPUT_DIR)
	rm -f pkg/collector/*.o
	rm -f pkg/collector/libsnapshot.a
	rm -f coverage.out coverage.html

## install: Install the exporter
install:
	go install $(LDFLAGS) ./cmd/exporter

## lint: Run golangci-lint
lint:
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

## fmt: Format code
fmt:
	gofmt -w -s .

## vet: Run go vet
vet:
	go vet ./...

## check: Run all checks (fmt, vet, lint, test)
check: fmt vet lint test
	@echo "All checks passed!"

## deps: Download and tidy dependencies
deps:
	go mod download
	go mod tidy

.DEFAULT_GOAL := help
