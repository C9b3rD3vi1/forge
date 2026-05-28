FROM golang:1.23-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o forge .

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata sqlite-libs

WORKDIR /app
COPY --from=builder /app/forge .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static

RUN mkdir -p /app/data /app/uploads

EXPOSE 3031

VOLUME ["/app/data", "/app/uploads"]

CMD ["./forge"]
