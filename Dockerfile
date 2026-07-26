FROM golang:alpine AS builder

ARG VERSION=dev

COPY . /app

WORKDIR /app

RUN  apk upgrade && apk add git npm && \
npm ci && npm run package && \
CGO_ENABLED=0 go build -ldflags "-s -w -X github.com/axllent/mailpit/config.Version=${VERSION}" -o /mailpit

FROM alpine:latest

LABEL org.opencontainers.image.title="Dispatch Inbox" \
  org.opencontainers.image.description="A self-hosted notification inbox for Dispatch, powered by Mailpit" \
  org.opencontainers.image.source="https://github.com/XxKotfeJxX/Dispatch-Inbox" \
  org.opencontainers.image.url="https://github.com/XxKotfeJxX/Dispatch-Inbox" \
  org.opencontainers.image.documentation="https://github.com/XxKotfeJxX/Dispatch-Inbox#readme" \
  org.opencontainers.image.licenses="MIT"

COPY --from=builder /mailpit /mailpit

RUN apk upgrade --no-cache && apk add --no-cache tzdata

EXPOSE 1025/tcp 1110/tcp 8025/tcp

HEALTHCHECK --interval=15s --start-period=10s --start-interval=1s CMD ["/mailpit", "readyz"]

ENTRYPOINT ["/mailpit"]
