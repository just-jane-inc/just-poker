FROM golang:1.25-alpine AS builder
ARG SWAG_VERSION=latest
ENV GOPRIVATE=github.com/just-jane-inc/*

RUN apk add --no-cache git

WORKDIR /build

RUN --mount=type=cache,target=/root/.cache/go-build \
    go install github.com/swaggo/swag/v2/cmd/swag@${SWAG_VERSION}

COPY go.mod go.sum ./

RUN --mount=type=secret,id=netrc,target=/root/.netrc,mode=0400 \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

COPY . .

RUN --mount=type=secret,id=netrc,target=/root/.netrc,mode=0400 \
    --mount=type=cache,target=/root/.cache/go-build \
    swag init \
      --v3.1 \
      --parseDependency \
      --parseInternal \
      --md ./additional-docs \
      -g ./main.go \
      -o ./docs \
      --ot json,yaml

RUN --mount=type=secret,id=netrc,target=/root/.netrc,mode=0400 \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux \
    go build -trimpath -ldflags="-s -w" -o /out/just-poker .

RUN mkdir -p /opt/var/log && chown 65532:65532 /opt/var/log

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /out/just-poker /just-poker
COPY --from=builder --chown=65532:65532 /opt/var/log /opt/var/log

# documentation port only
EXPOSE 8080

USER nonroot:nonroot
ENTRYPOINT ["/just-poker"]
