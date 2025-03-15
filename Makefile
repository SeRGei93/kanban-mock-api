APP_NAME = kanban-mock
BUILD_DIR = build

GOOS_MAC = darwin
GOOS_LINUX = linux
GOOS_WINDOWS = windows
GOARCH = amd64

LDFLAGS = -s -w
GCFLAGS =

build-linux: $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=$(GOOS_LINUX) GOARCH=$(GOARCH) go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-linux ./cmd/kanban/main.go

build-windows: $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=$(GOOS_WINDOWS) GOARCH=$(GOARCH) go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME).exe ./cmd/kanban/main.go

build-mac: $(BUILD_DIR)
	CGO_ENABLED=1 GOOS=$(GOOS_MAC) GOARCH=$(GOARCH) go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-mac ./cmd/kanban/main.go

clean:
	rm -rf $(BUILD_DIR)

$(BUILD_DIR):
	@mkdir -p $(BUILD_DIR)
