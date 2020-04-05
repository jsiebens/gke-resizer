FROM golang:1.13 AS builder

RUN apt-get -qq update && apt-get -yqq install upx

ENV GO111MODULE=on \
  CGO_ENABLED=0 \
  GOOS=linux \
  GOARCH=amd64

WORKDIR /src

COPY . .
RUN go build \
  -a \
  -trimpath \
  -ldflags "-s -w -extldflags '-static'" \
  -installsuffix cgo \
  -tags netgo \
  -o /bin/gkeresizer \
  .

RUN strip /bin/gkeresizer

RUN upx -q -9 /bin/gkeresizer




FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /bin/gkeresizer /bin/gkeresizer

ENV PORT 8080

ENTRYPOINT ["/bin/gkeresizer"]
