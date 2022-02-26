FROM golang:1.16 AS build

WORKDIR /go/src/github.com/danielunderwood/log2http
COPY *.go ./
COPY go.* ./
RUN go build -v ./...

FROM debian:11-slim
WORKDIR /app
COPY --from=build /go/src/github.com/danielunderwood/log2http/log2http /app
# Install ca-certificates so TLS can be verified
RUN apt-get update && apt-get install -y ca-certificates && apt-get clean && rm -rf /var/lib/apt/lists/*
CMD ["/app/log2http"]