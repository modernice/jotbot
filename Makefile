.PHONY: test
test:
	@find . -type d -name gen -print0 | xargs -0 rm -r
	@go test ./...

.PHONY: docs
docs:
	@./scripts/dogfood.sh

.PHONY: generate
generate:
	go generate ./...
