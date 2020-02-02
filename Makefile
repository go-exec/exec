all:

lint-local:
	@echo "Running linters"
	golangci-lint cache clean
	golangci-lint run -v ./...

lint:
	@echo "Running linters"
	docker run --rm -v $(PWD):/app -w /app golangci/golangci-lint:v1.21.0 golangci-lint run -v ./...

tests:
	@echo "Running tests"
	@mkdir -p artifacts
	go test -race -count=1 -cover -coverprofile=artifacts/coverage.out -v ./...

coverage: tests
	@echo "Running tests & coverage"
	go tool cover -html=artifacts/coverage.out -o artifacts/coverage.html
