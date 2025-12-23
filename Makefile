.PHONY: all build run clean

all: run

build:
	go build -o bin/api cmd/api

run: build
	go run cmd/api/main.go

clean:
	rm -rf bin