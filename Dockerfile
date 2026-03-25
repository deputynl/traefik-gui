# Stage 1: build the Vue frontend
# Run on the builder's native platform — just produces JS, architecture doesn't matter.
FROM --platform=$BUILDPLATFORM node:22-alpine AS frontend
WORKDIR /app/web
COPY web/package.json web/package-lock.json* ./
RUN npm install
COPY web/ .
RUN npm run build

# Stage 2: build the Go binary cross-compiled for the target platform.
# BUILDPLATFORM = builder's native arch (e.g. amd64)
# TARGETOS / TARGETARCH = injected by `docker buildx build --platform`
FROM --platform=$BUILDPLATFORM golang:1.22-alpine AS backend
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
COPY --from=frontend /app/web/dist ./web/dist
# Declare without defaults — buildx MUST supply these via --platform flag.
ARG TARGETOS
ARG TARGETARCH
RUN go mod tidy && \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w" -o traefik-gui .

# Stage 3: minimal final image for the target platform.
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=backend /app/traefik-gui .
EXPOSE 8888
ENTRYPOINT ["/app/traefik-gui"]
