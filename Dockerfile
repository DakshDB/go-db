FROM golang:1.22-alpine AS builder

# Set the current working directory.
RUN mkdir -p /go-db
WORKDIR /go-db

# install the modules 
COPY go.mod . 
COPY go.sum .
RUN go mod download
COPY . .

# Build executable binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o go-db-app app/main.go

# Remove unnecessary folders
RUN rm -Rf tools

# Build the final image
FROM alpine:3.20

# Set the current working directory.
WORKDIR /go-db
COPY --from=builder /go-db /go-db
COPY --from=builder /go-db/config.yml.example /go-db/config.yml

EXPOSE 8080
CMD ["./go-db-app"]
