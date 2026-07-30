FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY iam-service/go.mod iam-service/go.sum ./iam-service/
COPY proto/go.mod ./proto/
RUN cd iam-service && go mod download

COPY proto/ ./proto/
COPY iam-service/ ./iam-service/

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /iam-service \
    ./iam-service/cmd/api/main.go

FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

ENV TZ=Asia/Jakarta

COPY --from=builder /iam-service .

EXPOSE 3001 50051

CMD ["./iam-service"]
