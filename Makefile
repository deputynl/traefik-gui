REGISTRY  ?= ghcr.io/deputynl
IMAGE     ?= traefik-gui
TAG       ?= latest
# VERSION is read from web/package.json if not supplied (e.g. VERSION=1.2.0 make release)
VERSION   ?= $(shell node -p "require('./web/package.json').version" 2>/dev/null)

.PHONY: build build-arm64 release tag login setup-builder

# ── day-to-day ───────────────────────────────────────────────────────────────

## Build + push multi-arch image (amd64 + arm64) to ghcr.io and create a git tag
release:
	docker buildx build \
		--platform linux/amd64,linux/arm64 \
		--push \
		-t $(REGISTRY)/$(IMAGE):$(TAG) \
		-t $(REGISTRY)/$(IMAGE):$(VERSION) .
	@$(MAKE) tag

## Create and push a git tag for the current version (idempotent)
tag:
	@if git rev-parse "v$(VERSION)" >/dev/null 2>&1; then \
		echo "Tag v$(VERSION) already exists — skipping."; \
	else \
		git tag "v$(VERSION)" && git push origin "v$(VERSION)" && echo "Tagged v$(VERSION)"; \
	fi

## Build for the local machine only (no push, loads into docker daemon)
build:
	docker buildx build \
		--platform linux/amd64 \
		--load \
		-t $(IMAGE):$(TAG) .

## Build arm64 image and load into local docker daemon (for testing)
build-arm64:
	docker buildx build \
		--platform linux/arm64 \
		--load \
		-t $(IMAGE):$(TAG)-arm64 .

# ── one-time setup ───────────────────────────────────────────────────────────

## Log in to ghcr.io (run once; needs GITHUB_PAT env var or will prompt)
login:
	@echo "$(GITHUB_PAT)" | docker login ghcr.io -u deputynl --password-stdin 2>/dev/null || \
		docker login ghcr.io -u deputynl

## Create a buildx builder with cross-compilation support (run once per machine)
setup-builder:
	docker buildx create \
		--name multibuilder \
		--driver docker-container \
		--bootstrap \
		--use 2>/dev/null || docker buildx use multibuilder
	docker run --rm --privileged tonistiigi/binfmt --install all
	@echo "Builder ready. Test with: docker buildx ls"
