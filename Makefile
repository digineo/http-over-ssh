.PHONY: help
help: ## prints this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

.PHONY: dry-run
dry-run: ## prepares a release, but does not publish it
	goreleaser release --skip-publish

.PHONY: release
release: github_token ## release the current tag
	goreleaser release

.PHONY: snapshot
snapshot: github_token ## release HEAD as "$last_tag-next"
	goreleaser release --snapshot

.PHONY: clean
clean: ## removes the dist directory
	rm -rf ./dist/

.PHONY: github_token
github_token:
	@if [ -z "$$GITHUB_TOKEN" ]; then \
		echo >&2 "\033[1;31mMissing GITHUB_TOKEN env var\033[0m"; \
		exit 1; \
	fi
