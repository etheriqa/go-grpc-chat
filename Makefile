SERVER = build/server
CLIENT = build/client
BOT    = build/bot
SRCS = $(shell find . -name '*.go')

.PHONY: all build proto

all: build

build: $(SERVER) $(CLIENT) $(BOT)

proto:
	protoc --go_out=plugins=grpc:proto chat.proto

$(SERVER): $(shell find server -name '*.go')
	go build -o $(SERVER) github.com/etheriqa/go-grpc-chat/server

$(CLIENT): $(shell find client -name '*.go')
	go build -o $(CLIENT) github.com/etheriqa/go-grpc-chat/client

$(BOT): $(shell find bot -name '*.go')
	go build -o $(BOT) github.com/etheriqa/go-grpc-chat/bot
