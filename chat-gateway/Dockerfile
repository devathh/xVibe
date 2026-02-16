FROM golang:1.25.5-alpine AS builder    

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o xvbchat ./cmd/main.go

# ---
FROM alpine:latest

COPY --from=builder ./app/xvbchat ./
COPY --from=builder ./app/.env ./
COPY --from=builder ./app/Makefile ./
COPY --from=builder ./app/configs ./configs/

RUN apk add make
RUN wget -q -O /usr/local/bin/mkcert https://github.com/FiloSottile/mkcert/releases/latest/download/mkcert-v1.4.4-linux-amd64 \
    && chmod +x /usr/local/bin/mkcert

RUN make mkcertlocal

EXPOSE 7081

CMD ["./xvbchat"]