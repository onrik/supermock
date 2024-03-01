FROM golang:1.22 as builder

ADD ./ /src
WORKDIR /src
RUN go build -o /tmp/supermock

FROM debian:bookworm-slim

COPY --from=builder /tmp/supermock /usr/bin/supermock

ENTRYPOINT ["supermock"]
