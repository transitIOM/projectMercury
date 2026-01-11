FROM golang:1.25

WORKDIR /usr/src/app

# dependency caching
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN swag init -g internal/handlers/api.go
RUN go build -v -o /usr/local/bin/app ./cmd/api/main.go

CMD ["app"]
