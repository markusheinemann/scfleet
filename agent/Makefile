BINARY_NAME := agent
BUILD_DIR   := build
CMD_PATH    := ./cmd/agent

PLATFORMS := \
	linux/amd64 \
	linux/arm64 \
	darwin/amd64 \
	darwin/arm64 \
	windows/amd64 \
	windows/arm64

.PHONY: build test test-coverage cobertura-install lint lint-install install clean $(PLATFORMS)

build: $(PLATFORMS)

$(PLATFORMS):
	$(eval OS   := $(word 1, $(subst /, ,$@)))
	$(eval ARCH := $(word 2, $(subst /, ,$@)))
	$(eval EXT  := $(if $(filter windows,$(OS)),.exe,))
	GOOS=$(OS) GOARCH=$(ARCH) go build -o $(BUILD_DIR)/$(BINARY_NAME)-$(OS)-$(ARCH)$(EXT) $(CMD_PATH)

test:
	go test ./...

cobertura-install:
	@which gocover-cobertura > /dev/null 2>&1 || go install github.com/boumenot/gocover-cobertura@latest

test-coverage: cobertura-install
	mkdir -p $(BUILD_DIR)
	go test -coverprofile=$(BUILD_DIR)/coverage.out ./...
	go tool cover -func=$(BUILD_DIR)/coverage.out
	gocover-cobertura < $(BUILD_DIR)/coverage.out > $(BUILD_DIR)/coverage.xml

lint-install:
	@which golangci-lint > /dev/null 2>&1 || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

lint: lint-install
	golangci-lint run ./...

install:
	go install $(CMD_PATH)

clean:
	rm -rf $(BUILD_DIR)
