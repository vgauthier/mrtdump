BINARY_NAME = "mrtdump"
BINARY_DIR = "bin"
.PHONY: build run 

build:
	@echo "Building $(BINARY_NAME)..."
	GOARCH=arm64 GOOS=darwin go build -o $(BINARY_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/$(BINARY_NAME)/main.go
	GOARCH=amd64 GOOS=darwin go build -o $(BINARY_DIR)/$(BINARY_NAME)-darwin-intel ./cmd/$(BINARY_NAME)/main.go
	GOARCH=amd64 GOOS=linux go build -o $(BINARY_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/$(BINARY_NAME)/main.go

run: build
	@echo "Running $(BINARY_NAME)-darwin-arm64..."
	./$(BINARY_DIR)/$(BINARY_NAME)-darwin-arm64

test:
	@echo "Running tests..."
	go test ./... -v

coverage:
	@echo "Running tests with coverage..."
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

clean:
	rm -rf $(BINARY_DIR) *.out coverage.html