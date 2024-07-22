FROM golang:1.22 as builder

ADD ./ /src
WORKDIR /src
RUN go build -o /tmp/supermock ./cmd/supermock

FROM debian:bookworm-slim

COPY --from=builder /tmp/supermock /usr/bin/supermock

EXPOSE 8000

ENTRYPOINT ["supermock", "-listen", "0.0.0.0:8000"]
