export GO111MODULE=on

.PHONY: help
help: ## prints this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

.PHONY: dry-run
dry-run: ## prepares a release, but does not publish it
	goreleaser release --rm-dist --skip-publish

.PHONY: release
release: github_token ## release the current tag
	goreleaser release --rm-dist

.PHONY: snapshot
snapshot: github_token ## release HEAD as "$last_tag-next"
	goreleaser release --rm-dist --snapshot --skip-publish

.PHONY: clean
clean: ## removes the dist directory
	rm -rf ./dist/ coverage.out

.PHONY: github_token
github_token:
	@if [ -z "$$GITHUB_TOKEN" ]; then \
		echo >&2 "\033[1;31mMissing GITHUB_TOKEN env var\033[0m"; \
		exit 1; \
	fi


## testing

.PHONY: coverage.out
coverage.out:
	go test -race -covermode=atomic -coverprofile=$@ ./...

coverage.html: coverage.out
	go tool cover -html $< -o $@

.PHONY: test
test: coverage.out ## runs tests
