FROM golang:1.21-alpine AS build
WORKDIR /app
COPY . .
RUN go build -o /go-test-sarif ./cmd/main.go

FROM alpine:latest
COPY --from=build /go-test-sarif /go-test-sarif
COPY action/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
