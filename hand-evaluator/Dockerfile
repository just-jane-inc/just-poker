FROM alpine:3.20 AS cbuilder

RUN apk add --no-cache gcc musl-dev
WORKDIR /build
COPY lib .
RUN gcc -Ofast -lc allfive.c pokerlib.c -o just_eval

FROM golang:1.25-alpine AS gobuilder
WORKDIR /build
COPY . .
RUN go build .

FROM alpine
COPY --from=cbuilder /build/just_eval /just_eval
COPY --from=gobuilder /build/hand-evaluator /hand-evaluator

ENTRYPOINT ["/hand-evaluator"]
