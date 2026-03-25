REGISTRY  ?= ghcr.io/deputynl
IMAGE     ?= traefik-gui
TAG       ?= latest

# ── local builds ────────────────────────────────────────────────────────────
.PHONY: build build-arm64 push push-arm64 setup-builder

## Build for the current machine (amd64)
build:
	docker buildx build \
		--platform linux/amd64 \
		--load \
		-t $(IMAGE):$(TAG) .

## Build for Raspberry Pi (arm64) and load into local daemon
build-arm64:
	docker buildx build \
		--platform linux/arm64 \
		--load \
		-t $(IMAGE):$(TAG)-arm64 .

# ── registry workflow ────────────────────────────────────────────────────────
## Push amd64 image to the internal registry
push:
	docker buildx build \
		--platform linux/amd64 \
		--push \
		-t $(REGISTRY)/$(IMAGE):$(TAG) .

## Push arm64 image to the internal registry (for the Pi)
push-arm64:
	docker buildx build \
		--platform linux/arm64 \
		--push \
		-t $(REGISTRY)/$(IMAGE):$(TAG)-arm64 .

## Push a multi-arch manifest (amd64 + arm64) in one shot
push-multi:
	docker buildx build \
		--platform linux/amd64,linux/arm64 \
		--push \
		-t $(REGISTRY)/$(IMAGE):$(TAG) .

# ── one-time setup ───────────────────────────────────────────────────────────
## Create a dedicated buildx builder that supports cross-compilation
setup-builder:
	docker buildx create \
		--name multibuilder \
		--driver docker-container \
		--bootstrap \
		--use
	docker run --rm --privileged tonistiigi/binfmt --install all
	@echo "Builder ready. Test with: docker buildx ls"
