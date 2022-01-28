# Use the offical golang image to create a binary.
FROM golang:1.17 AS builder

# Disable CGO to get a static go binary.
ENV CGO_ENABLED=0

# Create and change to the app directory.
WORKDIR /go/src/app

# Copy local code to the container image.
COPY . .

# Build a single static binary.
#   - "a" recompile symbols for our production build
#   - "trimpath" makes stacktraces nicer and gets a reproducible build
#   - "ldflags" strip the binary and tells it to compile statically
RUN go build \
  -a \
  -trimpath \
  -ldflags "-s -w -extldflags '-static'" \
  -o /go/bin/app \
  ./test/grpc-app

# Strip symbols from binary to make it smaller.
RUN strip -s /go/bin/app

# Create the user `nobody`.
RUN echo "nobody:*:65534:65534:nobody:/:/bin/false" > /tmp/etc-passwd

# Use a scratch image to host our binary.
FROM scratch
COPY --from=builder /tmp/etc-passwd /etc/passwd
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/app /app
USER nobody

# Run the web service on container startup.
ENTRYPOINT ["/app"]