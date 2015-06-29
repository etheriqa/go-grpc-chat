SERVER = srv
CLIENT = cli
SRCS = $(shell find . -name '*.go')

.PHONY: all build proto

all: build

build: $(SERVER) $(CLIENT)

proto:
	protoc --go_out=plugins=grpc:proto chat.proto

$(SERVER): $(SRCS)
	go build -o $(SERVER) github.com/etheriqa/go-grpc-chat/server

$(CLIENT): $(SRCS)
	go build -o $(CLIENT) github.com/etheriqa/go-grpc-chat/client
