FROM golang:1.25-alpine AS builder

COPY core-service/ /app/core-service/
COPY proto/ /app/proto/

WORKDIR /app/core-service
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /core-service \
    ./cmd/api/main.go

FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Jakarta

COPY --from=builder /core-service .

EXPOSE 3002
CMD ["./core-service"]
