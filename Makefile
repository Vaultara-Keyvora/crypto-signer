all: build

build:
	go build -o bin/crypto_signer cmd/active/main.go