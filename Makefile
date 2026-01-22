.PHONY: build build-ui build-backend run clean docker-build docker-push test

# Variables
APP_NAME := middleware-manager
DOCKER_REPO := hhftechnology
DOCKER_TAG := latest
GO_FILES := $(shell find . -name "*.go" -not -path "./vendor/*")

# Default target
all: build

# Build everything
build: build-ui build-backend

# Build UI (Vite + TypeScript)
build-ui:
	@echo "Building UI..."
	cd ui && npm install && npm run build

# Build backend
build-backend:
	@echo "Building backend..."
	go build -o $(APP_NAME) .

# Run the application
run: build
	@echo "Running application..."
	./$(APP_NAME)

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(APP_NAME)
	rm -rf ui/dist

# Build Docker image
docker-build: build
	@echo "Building Docker image..."
	docker build -t $(DOCKER_REPO)/$(APP_NAME):$(DOCKER_TAG) .

# Push Docker image
docker-push: docker-build
	@echo "Pushing Docker image..."
	docker push $(DOCKER_REPO)/$(APP_NAME):$(DOCKER_TAG)

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run the application in development mode
dev:
	@echo "Running in development mode..."
	go run main.go

# Run the UI in development mode
dev-ui:
	@echo "Running UI in development mode..."
	cd ui && npm run dev

# Install dependencies
deps:
	@echo "Installing Go dependencies..."
	go mod download
	@echo "Installing UI dependencies..."
	cd ui && npm install

# Type check the UI
typecheck:
	@echo "Type checking UI..."
	cd ui && npm run typecheck

# Lint the UI
lint-ui:
	@echo "Linting UI..."
	cd ui && npm run lint
