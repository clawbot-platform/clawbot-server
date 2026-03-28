FROM golang:1.26-alpine AS build

WORKDIR /src
RUN apk add --no-cache ca-certificates git

ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/

RUN go build -ldflags "-X 'clawbot-server/internal/version.Value=${VERSION}' -X 'clawbot-server/internal/version.Commit=${COMMIT}' -X 'clawbot-server/internal/version.BuildDate=${BUILD_DATE}'" -o /out/clawbot-server ./cmd/clawbot-server

FROM alpine:3.21

ARG OCI_SOURCE="https://github.com/clawbot-platform/clawbot-server"
ARG OCI_DESCRIPTION="Reusable Go-first control-plane foundation for Clawbot Platform services."
ARG OCI_LICENSES="Apache-2.0"

LABEL org.opencontainers.image.source="${OCI_SOURCE}" \
      org.opencontainers.image.description="${OCI_DESCRIPTION}" \
      org.opencontainers.image.licenses="${OCI_LICENSES}"

RUN apk add --no-cache ca-certificates wget && adduser -D -u 10001 clawbot
COPY --from=build /out/clawbot-server /usr/local/bin/clawbot-server

USER clawbot
WORKDIR /app
EXPOSE 8080

ENTRYPOINT ["clawbot-server", "serve"]
