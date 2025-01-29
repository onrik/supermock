FROM golang:1.23 AS builder

ADD ./ /src
WORKDIR /src
RUN go build -o /tmp/supermock ./cmd/supermock

FROM debian:bookworm-slim

COPY --from=builder /tmp/supermock /usr/bin/supermock

EXPOSE 8000

ENTRYPOINT ["supermock"]
