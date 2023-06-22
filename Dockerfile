# Build stage
FROM golang:1.20.2-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o main main.go

# Run stage
FROM alpine
ENV TZ=America/Sao_Paulo
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/main .
COPY .env .
COPY ./dbc/dbcParaDbf ./dbc/dbcParaDbf

CMD [ "/app/main" ]