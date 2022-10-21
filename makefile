test:
	@echo "Testing all features..."

tidy:
	@echo "Installing project dependencies"
	@go mod tidy

release:
	@echo "Firing new release helper"
	@npx release-it