FROM --platform=$BUILDPLATFORM golang:1.26-alpine3.23 AS build

RUN apk add --no-cache make g++ nodejs npm

WORKDIR /PrivUtil

# Cache Go module downloads separately from source changes
COPY go.mod go.sum ./
RUN go mod download

# Cache npm installs separately from source changes
COPY web/package.json web/package-lock.json ./web/
RUN cd web && npm ci

# Copy source (node_modules excluded via .dockerignore)
COPY . .

# Version is injected by CI; the build context has no .git for `git describe`,
# so pass it through to the ldflags via the Makefile (overrides the git-derived default).
# Declared here (after the cached layers) so changing it doesn't bust the module/npm caches.
ARG VERSION=dev
# TARGETOS/TARGETARCH are supplied by buildx per target platform. The build stage
# runs on the native BUILDPLATFORM (the frontend build is arch-independent) and Go
# cross-compiles (CGO-free) to the target, so multi-arch images build without
# emulating the whole toolchain.
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} make build-go BUILD_VERSION=${VERSION}

FROM alpine:3.23
RUN addgroup -S privutil && adduser -S privutil -G privutil
COPY --from=build /PrivUtil/privutil /bin/privutil
USER privutil
ENTRYPOINT ["privutil"]
