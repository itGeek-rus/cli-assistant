FROM golang:1.26-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
ARG VERSION=dev
ARG COMMIT=none
ARG BUILD_DATE=unknown
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags "\
    -s -w\
    -X cli-assistant/pkg/version.Version=${VERSION} \
    -X cli-assistant/pkg/version.Commit=${COMMIT} \
    -X cli-assistant/pkg/version.BuildDate=${BUILD_DATE}" \
    -o /out/assistant ./cmd/app

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /out/assistant /usr/local/bin/assistant

ENTRYPOINT ["/usr/local/bin/assistant"]

LABEL authors="vacheslavterentev"