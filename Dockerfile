# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang:alpine AS build

# Copy the local package files to the container's workspace.
WORKDIR /app
COPY . .

# Build the app command inside the container.
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /go-todo-list
RUN ls

# Expose port 8080
EXPOSE 8080

RUN ls

# Start from a scratch image for the runtime stage
FROM scratch

WORKDIR /

# Copy the app binary from the build stage
COPY --from=build /go-todo-list /go-todo-list


# Expose port 8080
EXPOSE 8080

# Run the app command by default when the container starts.
ENTRYPOINT ["/go-todo-list"]
