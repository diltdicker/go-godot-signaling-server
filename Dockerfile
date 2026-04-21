# Stage 1: The Build Environment
FROM golang:alpine AS builder

WORKDIR /app

ARG GO_DIR=go

COPY ${GO_DIR} .

RUN CGO_ENABLED=0 GOOS=linux go build -o rtc_server .

# Stage 2: The Final Image
FROM alpine:latest

# setup non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /root/

COPY --from=builder /app/rtc_server .

RUN chown appuser:appgroup rtc_server

USER appuser

EXPOSE 8080

CMD ["./rtc_server"]