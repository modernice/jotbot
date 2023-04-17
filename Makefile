.PHONY: test
test:
	@go test ./...

.PHONY: clear
clear:
	@./scripts/clear.sh

.PHONY: docs
docs:
	@./scripts/dogfood.sh
