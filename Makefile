.PHONY: test
test:
	@find . -type d -name gen -print0 | xargs -0 rm -r
	@go test ./...

.PHONY: clear
clear:
	@./scripts/clear.sh

.PHONY: docs
docs: clear
	@./scripts/dogfood.sh
