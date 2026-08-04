FROM golang:1.25-alpine AS builder
ARG SWAG_VERSION=latest

WORKDIR /build

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go install github.com/swaggo/swag/v2/cmd/swag@${SWAG_VERSION}

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    swag init \
      --v3.1 \
      --parseDependency \
      --parseInternal \
      --md ./additional-docs \
      -g ./main.go \
      -o ./docs \
      --ot json,yaml

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux \
    go build -trimpath -ldflags="-s -w" -o /out/just-poker .

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /out/just-poker /just-poker

# documentation port only
EXPOSE 8080

USER nonroot:nonroot
ENTRYPOINT ["/just-poker"]
