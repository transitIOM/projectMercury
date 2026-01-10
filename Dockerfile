FROM golang:1.24.3

WORKDIR /usr/src/app

# dependency caching
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN make docs
RUN go build -v -o /usr/local/bin/app ./cmd/api/main.go

CMD ["app"]
