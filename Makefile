.PHONY: test
test:
	@go test ./...
	@find . -type d -name gen -delete

.PHONY: clear
clear:
	@./scripts/clear.sh

.PHONY: docs
docs: clear
	@./scripts/dogfood.sh
