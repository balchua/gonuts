# Use base golang image from Docker Hub
FROM golang:1.15.3 as build

WORKDIR /src/github.com/balchua/gonuts

# Copy go.mod and go.sum 
ADD ./go.mod /src/github.com/balchua/gonuts/
ADD ./go.sum /src/github.com/balchua/gonuts/
# Install dependencies in go.mod and go.sum
RUN go mod download

# Copy application source code
COPY ./main.go /src/github.com/balchua/gonuts

# Compile the application to /app.
RUN go build -o /app -v .

# Now create separate deployment image
FROM gcr.io/distroless/base
COPY --from=build /app /app
# Cause full tracebacks; also serves to identify this image as a Go image for `skaffold debug`
ENV GOTRACEBACK=all
ENTRYPOINT ["/app"]
