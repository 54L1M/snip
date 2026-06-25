# --- Variables ---
BINARY_NAME=snip
CMD_PATH=./main.go
VERSION_VAR_PATH="github.com/54L1M/snip/cmd.version"

LOCAL_VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "(devel)")
LOCAL_LDFLAGS=-ldflags "-X '$(VERSION_VAR_PATH)=$(LOCAL_VERSION)'"

.PHONY: all build run clean test vet help install install-config

help: # Show help for each of the Makefile recipes.
	@grep -E '^[a-zA-Z0-9 -]+:.*#'  Makefile | sort | while read -r l; do printf "\033[1;32m$$(echo $$l | cut -f 1 -d':')\033[00m:$$(echo $$l | cut -f 2- -d'#')\n"; done

build: # Compiles the application from your CURRENT local files.
	@echo "Building binary from local source (version: $(LOCAL_VERSION))..."
	@mkdir -p bin
	@go build $(LOCAL_LDFLAGS) -o bin/$(BINARY_NAME) $(CMD_PATH)

run: build # Compiles and runs the application from your CURRENT local files.
	@./bin/$(BINARY_NAME) $(ARGS)

vet: # Runs go vet across the module.
	@go vet ./...

test: # Runs all tests in verbose mode.
	@echo "Running tests..."
	@go test -v ./...

clean: # Removes the compiled binary and the bin directory.
	@echo "Cleaning up..."
	@rm -rf bin

install: # Builds and installs the LATEST git tag version locally. ✨
	@echo "Fetching the latest tags from the remote..."
	@git fetch --tags
	@CURRENT_BRANCH=$$(git rev-parse --abbrev-ref HEAD); \
	LATEST_TAG=$$(git tag --sort=-v:refname | head -n 1); \
	if [ -z "$$LATEST_TAG" ]; then \
		echo "Error: No git tags found. Please create a tag first."; \
		exit 1; \
	fi; \
	echo "Found latest tag: $$LATEST_TAG"; \
	git checkout -q $$LATEST_TAG; \
	mkdir -p bin; \
	go build -ldflags "-X '$(VERSION_VAR_PATH)=$$LATEST_TAG'" -o bin/$(BINARY_NAME) $(CMD_PATH); \
	echo "Installing $(BINARY_NAME) (version: $$LATEST_TAG) to /usr/local/bin..."; \
	sudo cp ./bin/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME); \
	git checkout -q $$CURRENT_BRANCH; \
	echo "✅ Installation complete."

install-config: # Copies the example config file to the user's config directory.
	@echo "Installing example config to ~/.config/snip/.snip..."
	@mkdir -p ~/.config/snip
	@cp ./cmd/.snip.example ~/.config/snip/.snip
