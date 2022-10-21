test:
	@echo "Testing all features..."

tidy:
	@echo "Installing project dependencies"
	@go mod tidy

release:
	@echo "Firing new release helper"
	@npx release-it

prepare: tidy
	@echo "installing goose binary"
	@go install github.com/pressly/goose/v3/cmd/goose@latest