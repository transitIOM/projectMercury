BIN_NAME := api
CMD_PKG := ./cmd/api

.PHONY: all build run clean

all: run

build:
	go build -o bin/$(BIN_NAME) $(CMD_PKG)

run: build
	./bin/$(BIN_NAME)

clean:
	rm -rf bin