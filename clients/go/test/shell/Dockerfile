FROM --platform=$BUILDPLATFORM golang:latest AS builder

ENV PORT=8080
ENV CGO_ENABLED=0
ENV GOPROXY=https://proxy.golang.org,direct

WORKDIR /go/src/app
COPY . .

RUN go build \
  -a \
  -trimpath \
  -ldflags "-s -w -extldflags='-static'" \
  -o /go/bin/app \
  ./clients/go/test/shell

RUN strip -s /go/bin/app

RUN echo "nobody:*:65534:65534:nobody:/:/bin/false" > /tmp/etc-passwd

# Use a scratch image to host our binary.
FROM scratch
COPY --from=builder /tmp/etc-passwd /etc/passwd
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/app /app
# Copy the client config.
COPY clients/go/test/shell/config.yaml /etc/lumberjack/config.yaml
COPY integration/testrunner/test_jwks test_jwks

USER nobody

EXPOSE 8080
ENTRYPOINT ["/app"]
