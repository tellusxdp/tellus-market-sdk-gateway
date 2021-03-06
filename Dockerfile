# Builder
FROM golang:1.13 as builder

ENV CGO_ENABLED=0

WORKDIR /go/src/github.com/tellusxdp/tellus-market-sdk-gateway
ADD . .

RUN go get
RUN go build main.go


# Image
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/github.com/tellusxdp/tellus-market-sdk-gateway/main /opt/bin/market-gateway
CMD ["/opt/bin/market-gateway", "--config", "/opt/market/config.yml"]

