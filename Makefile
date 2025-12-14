.PHONY: build run test clean init deps help

# Имя приложения
APP_NAME := local-video-server

# Версия
VERSION := 0.1.0

# Go параметры
GO := go
GOFLAGS := -ldflags="-X main.version=$(VERSION)"

help: ## Показать справку
	@echo "Доступные команды:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

deps: ## Установить зависимости
	$(GO) mod download
	$(GO) mod tidy

build: ## Собрать приложение
	$(GO) build $(GOFLAGS) -o bin/$(APP_NAME) ./cmd/server

build-windows: ## Собрать для Windows
	GOOS=windows GOARCH=amd64 $(GO) build $(GOFLAGS) -o bin/$(APP_NAME).exe ./cmd/server

build-linux: ## Собрать для Linux
	GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -o bin/$(APP_NAME)-linux ./cmd/server

build-darwin: ## Собрать для macOS
	GOOS=darwin GOARCH=amd64 $(GO) build $(GOFLAGS) -o bin/$(APP_NAME)-darwin ./cmd/server

build-all: build-windows build-linux build-darwin ## Собрать для всех платформ

run: build ## Собрать и запустить
	./bin/$(APP_NAME)

test: ## Запустить тесты
	$(GO) test -v ./...

clean: ## Очистить артефакты сборки
	rm -rf bin/
	$(GO) clean

init: ## Инициализировать проект (создать конфиг)
	$(GO) run ./cmd/server init

fmt: ## Форматировать код
	$(GO) fmt ./...

vet: ## Проверить код
	$(GO) vet ./...

lint: fmt vet ## Запустить линтер и форматирование

install: build ## Установить приложение
	cp bin/$(APP_NAME) /usr/local/bin/$(APP_NAME)
