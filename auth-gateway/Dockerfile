FROM golang:1.25.5-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o xvb ./cmd/main.go

# ---
FROM alpine:latest  

COPY --from=builder ./app/xvb ./
COPY --from=builder ./app/.env ./
COPY --from=builder ./app/configs ./configs/
COPY --from=builder ./app/certs ./certs/

CMD [ "./xvb" ]
