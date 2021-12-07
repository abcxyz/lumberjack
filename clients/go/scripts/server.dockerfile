FROM golang:1.17 AS builder

WORKDIR /go/src/app
COPY . .

ENV CGO_ENABLED=0
RUN go build \
  -a \
  -trimpath \
  -ldflags "-s -w -extldflags '-static'" \
  -o /go/bin/server \
  ./cmd/server
RUN strip -s /go/bin/server

RUN echo "nobody:*:65534:65534:nobody:/:/bin/false" > /tmp/etc-passwd

# Use a scratch image to host our binary.
FROM scratch
COPY --from=builder /tmp/etc-passwd /etc/passwd
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/server /server
USER nobody

ENV PORT 8080
ENTRYPOINT ["/server"]