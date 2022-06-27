# Step 1: Builder
FROM golang:1.18.3 as builder
COPY . /app
WORKDIR /app
RUN CGO_ENABLED=0 go build -o /app/bin/app ./cmd/main.go

# Step 2: Create image
FROM alpine:latest
WORKDIR /app
EXPOSE 8080/tcp
EXPOSE 8081/tcp
COPY --from=builder /app/bin/app /app/bin/app
CMD ["/app/bin/app", "-hsm-addr", "host.docker.internal:1500"]