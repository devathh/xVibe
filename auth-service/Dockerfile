FROM golang:1.25.5-alpine AS builder    

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o xvbauth ./cmd/main.go

# ---
FROM alpine:latest

COPY --from=builder ./app/xvbauth ./
COPY --from=builder ./app/.env ./
COPY --from=builder ./app/Makefile ./
COPY --from=builder ./app/configs ./configs/

RUN apk add make
RUN apk add --no-cache ca-certificates openssl
RUN wget -q -O /usr/local/bin/mkcert https://github.com/FiloSottile/mkcert/releases/latest/download/mkcert-v1.4.4-linux-amd64 \
    && chmod +x /usr/local/bin/mkcert

RUN make keys
RUN make mtls

EXPOSE 50050

CMD ["./xvbauth"]