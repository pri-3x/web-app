# Build stage
FROM golang:1.22.7-alpine AS build

WORKDIR /app

COPY go.mod ./
COPY . .

RUN go build -o /go-web-app ./cmd/server

# Run stage
FROM alpine:latest

WORKDIR /

COPY --from=build /go-web-app /go-web-app

EXPOSE 8085

CMD ["/go-web-app"]