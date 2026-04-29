FROM golang:1.26-alpine3.23 AS build

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

RUN make build-go

FROM alpine:3.23
RUN addgroup -S privutil && adduser -S privutil -G privutil
COPY --from=build /PrivUtil/privutil /bin/privutil
USER privutil
ENTRYPOINT ["privutil"]
