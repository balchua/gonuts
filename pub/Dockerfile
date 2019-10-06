# Use base golang image from Docker Hub
FROM golang:1.12.10 as build

WORKDIR /src/github.com/balchua/gonuts

# Copy application source code
COPY ./ /src/github.com/balchua/gonuts
# Install dependencies in go.mod and go.sum
RUN go mod download

# Compile the application to /app.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64  go build -o /app -v .

# Now create separate deployment image
FROM gcr.io/distroless/base:debug
COPY --from=build /app /app
# Cause full tracebacks; also serves to identify this image as a Go image for `skaffold debug`
ENV GOTRACEBACK=all
ENTRYPOINT ["/app", "-s", "nats://stan-nats-ss.stan-nats-ss.stan.svc.cluster.local:4222", "-id","sample-pub", "Test"]
