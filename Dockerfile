FROM golang:1.24-alpine AS build
RUN useradd -m appuser
USER appuser
WORKDIR /app
COPY . .
RUN go build -o /go-test-sarif ./cmd/main.go

FROM alpine:3.21.3
RUN useradd -m appuser
USER appuser
COPY --from=build /go-test-sarif /go-test-sarif
COPY action/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
