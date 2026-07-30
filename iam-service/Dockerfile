FROM golang:1.25-alpine AS builder

COPY iam-service/ /app/iam-service/
COPY proto/ /app/proto/

WORKDIR /app/iam-service
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /iam-service \
    ./cmd/api/main.go

FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Jakarta

COPY --from=builder /iam-service .

EXPOSE 3001 50051
CMD ["./iam-service"]
