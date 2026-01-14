.PHONY: build run test clean install lint fmt vet check deps help package-all package-tar package-rpm package-deb

BINARY_NAME=faircom_exporter
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "1.0.0")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT_SHA=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS=-ldflags "-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.CommitSHA=$(COMMIT_SHA)"
OUTPUT_DIR=output
PACKAGE_NAME=$(BINARY_NAME)-$(VERSION)

## help: Display this help message
help:
	@echo "Available targets:"
	@awk '/^##/ {printf "\033[36m%-20s\033[0m %s\n", $$2, substr($$0, index($$0, $$3))}' $(MAKEFILE_LIST)

## build: Build the exporter binary for current platform
build:
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(OUTPUT_DIR)/$(BINARY_NAME) ./cmd/exporter

## build-linux-amd64: Build for Linux AMD64
build-linux-amd64:
	mkdir -p $(OUTPUT_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(OUTPUT_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/exporter

## build-linux-arm64: Build for Linux ARM64
build-linux-arm64:
	mkdir -p $(OUTPUT_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(OUTPUT_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/exporter

## package-tar: Create tar.gz packages for Linux
package-tar: build-linux-amd64 build-linux-arm64
	mkdir -p $(OUTPUT_DIR)/tar
	# AMD64 package
	mkdir -p $(OUTPUT_DIR)/tar/tmp-amd64/$(PACKAGE_NAME)-linux-amd64
	cp $(OUTPUT_DIR)/$(BINARY_NAME)-linux-amd64 $(OUTPUT_DIR)/tar/tmp-amd64/$(PACKAGE_NAME)-linux-amd64/$(BINARY_NAME)
	cp config.yaml $(OUTPUT_DIR)/tar/tmp-amd64/$(PACKAGE_NAME)-linux-amd64/
	cp README.md $(OUTPUT_DIR)/tar/tmp-amd64/$(PACKAGE_NAME)-linux-amd64/
	tar -czf $(OUTPUT_DIR)/tar/$(PACKAGE_NAME)-linux-amd64.tar.gz -C $(OUTPUT_DIR)/tar/tmp-amd64 $(PACKAGE_NAME)-linux-amd64
	rm -rf $(OUTPUT_DIR)/tar/tmp-amd64
	# ARM64 package
	mkdir -p $(OUTPUT_DIR)/tar/tmp-arm64/$(PACKAGE_NAME)-linux-arm64
	cp $(OUTPUT_DIR)/$(BINARY_NAME)-linux-arm64 $(OUTPUT_DIR)/tar/tmp-arm64/$(PACKAGE_NAME)-linux-arm64/$(BINARY_NAME)
	cp config.yaml $(OUTPUT_DIR)/tar/tmp-arm64/$(PACKAGE_NAME)-linux-arm64/
	cp README.md $(OUTPUT_DIR)/tar/tmp-arm64/$(PACKAGE_NAME)-linux-arm64/
	tar -czf $(OUTPUT_DIR)/tar/$(PACKAGE_NAME)-linux-arm64.tar.gz -C $(OUTPUT_DIR)/tar/tmp-arm64 $(PACKAGE_NAME)-linux-arm64
	rm -rf $(OUTPUT_DIR)/tar/tmp-arm64
	@echo "Tar packages created in $(OUTPUT_DIR)/tar/"

## package-deb: Create DEB packages for Debian/Ubuntu
package-deb: build-linux-amd64 build-linux-arm64
	@which fpm > /dev/null || (echo "ERROR: fpm not found. Install with: gem install fpm" && exit 1)
	mkdir -p $(OUTPUT_DIR)/deb
	# AMD64 package
	fpm -s dir -t deb -n $(BINARY_NAME) -v $(VERSION) \
		-a amd64 \
		--description "Prometheus exporter for FairCom database metrics" \
		--url "https://github.com/faircom/prometheus-exporter" \
		--license "MIT" \
		--maintainer "FairCom" \
		--deb-systemd systemd/$(BINARY_NAME).service \
		--config-files /etc/$(BINARY_NAME)/config.yaml \
		--directories /etc/$(BINARY_NAME) \
		--package $(OUTPUT_DIR)/deb/$(PACKAGE_NAME)-amd64.deb \
		$(OUTPUT_DIR)/$(BINARY_NAME)-linux-amd64=/usr/local/bin/$(BINARY_NAME) \
		config.yaml=/etc/$(BINARY_NAME)/config.yaml
	# ARM64 package
	fpm -s dir -t deb -n $(BINARY_NAME) -v $(VERSION) \
		-a arm64 \
		--description "Prometheus exporter for FairCom database metrics" \
		--url "https://github.com/faircom/prometheus-exporter" \
		--license "MIT" \
		--maintainer "FairCom" \
		--deb-systemd systemd/$(BINARY_NAME).service \
		--config-files /etc/$(BINARY_NAME)/config.yaml \
		--directories /etc/$(BINARY_NAME) \
		--package $(OUTPUT_DIR)/deb/$(PACKAGE_NAME)-arm64.deb \
		$(OUTPUT_DIR)/$(BINARY_NAME)-linux-arm64=/usr/local/bin/$(BINARY_NAME) \
		config.yaml=/etc/$(BINARY_NAME)/config.yaml
	@echo "DEB packages created in $(OUTPUT_DIR)/deb/"

## package-rpm: Create RPM packages for RHEL/CentOS
package-rpm: build-linux-amd64 build-linux-arm64
	@which fpm > /dev/null || (echo "ERROR: fpm not found. Install with: gem install fpm" && exit 1)
	mkdir -p $(OUTPUT_DIR)/rpm
	# AMD64 package
	fpm -s dir -t rpm -n $(BINARY_NAME) -v $(VERSION) \
		-a x86_64 \
		--description "Prometheus exporter for FairCom database metrics" \
		--url "https://github.com/faircom/prometheus-exporter" \
		--license "MIT" \
		--maintainer "FairCom" \
		--config-files /etc/$(BINARY_NAME)/config.yaml \
		--directories /etc/$(BINARY_NAME) \
		--package $(OUTPUT_DIR)/rpm/$(PACKAGE_NAME)-amd64.rpm \
		$(OUTPUT_DIR)/$(BINARY_NAME)-linux-amd64=/usr/local/bin/$(BINARY_NAME) \
		config.yaml=/etc/$(BINARY_NAME)/config.yaml \
		systemd/$(BINARY_NAME).service=/usr/lib/systemd/system/$(BINARY_NAME).service
	# ARM64 package
	fpm -s dir -t rpm -n $(BINARY_NAME) -v $(VERSION) \
		-a aarch64 \
		--description "Prometheus exporter for FairCom database metrics" \
		--url "https://github.com/faircom/prometheus-exporter" \
		--license "MIT" \
		--maintainer "FairCom" \
		--config-files /etc/$(BINARY_NAME)/config.yaml \
		--directories /etc/$(BINARY_NAME) \
		--package $(OUTPUT_DIR)/rpm/$(PACKAGE_NAME)-arm64.rpm \
		$(OUTPUT_DIR)/$(BINARY_NAME)-linux-arm64=/usr/local/bin/$(BINARY_NAME) \
		config.yaml=/etc/$(BINARY_NAME)/config.yaml \
		systemd/$(BINARY_NAME).service=/usr/lib/systemd/system/$(BINARY_NAME).service
	@echo "RPM packages created in $(OUTPUT_DIR)/rpm/"

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
	rm -f coverage.out coverage.html

## install: Install the exporter
install:
	go install $(LDFLAGS) ./cmd/exporter

## build-all: Build for all platforms
build-all:
	mkdir -p $(OUTPUT_DIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(OUTPUT_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/exporter
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(OUTPUT_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/exporter
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(OUTPUT_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/exporter
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(OUTPUT_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/exporter
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(OUTPUT_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/exporter

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
