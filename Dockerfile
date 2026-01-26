# ---- build stage ----
FROM golang:1.25 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/api

# ---- run stage ----
FROM gcr.io/distroless/base-debian12
WORKDIR /
COPY --from=builder /server /server

EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/server"]
