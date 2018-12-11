# https://goreleaser.com/

dry-run:
	goreleaser release --skip-publish

release: github_token
	goreleaser release

.PHONY: github_token
github_token:
	@if [ -z "$$GITHUB_TOKEN" ]; then \
		echo >&2 "\033[1;31mMissing GITHUB_TOKEN env var\033[0m"; \
		exit 1; \
	fi
