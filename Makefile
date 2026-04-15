.PHONY: build build-ui build-traefik-ui build-backend run clean docker-build docker-push test dev dev-ui dev-traefik-ui dev-traefik-manager deps typecheck typecheck-traefik-ui lint-ui

APP_NAME := middleware-manager
DOCKER_REPO := hhftechnology
DOCKER_TAG := latest
GO_FILES := $(shell find . -name "*.go" -not -path "./vendor/*")

all: build

build: build-ui build-traefik-ui build-backend

build-ui:
	@echo "Building middleware-manager UI..."
	cd ui && npm install && npm run build

build-traefik-ui:
	@echo "Building Traefik Manager UI..."
	cd traefik-ui && npm install && npm run build

build-backend:
	@echo "Building backend..."
	go build -o $(APP_NAME) .

run: build
	@echo "Running application..."
	./$(APP_NAME)

clean:
	@echo "Cleaning..."
	rm -f $(APP_NAME)
	rm -rf ui/dist traefik-ui/dist

docker-build: build
	@echo "Building Docker image..."
	docker build -t $(DOCKER_REPO)/$(APP_NAME):$(DOCKER_TAG) .

docker-push: docker-build
	@echo "Pushing Docker image..."
	docker push $(DOCKER_REPO)/$(APP_NAME):$(DOCKER_TAG)

test:
	@echo "Running tests..."
	go test -v ./...

dev:
	@echo "Running middleware-manager in development mode..."
	go run .

dev-ui:
	@echo "Running middleware-manager UI in development mode..."
	cd ui && npm run dev

dev-traefik-ui:
	@echo "Running Traefik Manager UI in development mode..."
	cd traefik-ui && npm run dev

dev-traefik-manager:
	@echo "Running Traefik Manager backend in development mode..."
	MODE=traefik-manager go run .

deps:
	@echo "Installing Go dependencies..."
	go mod download
	@echo "Installing middleware-manager UI dependencies..."
	cd ui && npm install
	@echo "Installing Traefik Manager UI dependencies..."
	cd traefik-ui && npm install

typecheck:
	@echo "Type checking middleware-manager UI..."
	cd ui && npm run typecheck

typecheck-traefik-ui:
	@echo "Type checking Traefik Manager UI..."
	cd traefik-ui && npm run typecheck

lint-ui:
	@echo "Linting middleware-manager UI..."
	cd ui && npm run lint
