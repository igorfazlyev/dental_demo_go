FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY . .
RUN go build -o dental-ai-platform main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/dental-ai-platform .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static

EXPOSE 8080
CMD ["./dental-ai-platform"]
