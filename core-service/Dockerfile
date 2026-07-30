FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY core-service/go.mod core-service/go.sum ./core-service/
COPY proto/go.mod ./proto/
RUN cd core-service && go mod download

COPY proto/ ./proto/
COPY core-service/ ./core-service/

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /core-service \
    ./core-service/cmd/api/main.go

FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

ENV TZ=Asia/Jakarta

COPY --from=builder /core-service .

EXPOSE 3002

CMD ["./core-service"]
