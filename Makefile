.PHONY: start build

NOW = $(shell date -u '+%Y%m%d%I%M%S')

all: start

build:
	@go build -ldflags "-w -s" -o ./cmd/phone-area main.go

build-win32:
	@go build -ldflags "-w -s" -o ./cmd/phone-area.exe main.go

start:
	go run main.go txt -f phone.txt