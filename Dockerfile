FROM golang:1.27.3-alpine3.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN GOTOOLCHAIN=local go mod tidy && CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o collector ./cmd/collector


FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /app/collector /collector

ENTRYPOINT ["/collector"]
