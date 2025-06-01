FROM golang:1.24-alpine

WORKDIR /app

RUN apk add --no-cache git bash curl

RUN go install github.com/air-verse/air@latest

COPY go.*  ./
RUN go mod download

COPY . .

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]