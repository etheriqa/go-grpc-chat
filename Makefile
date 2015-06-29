SERVER = build/server
CLIENT = build/client
BOT    = build/bot
SERVER_SRCS = $(shell find server -name '*.go')
CLIENT_SRCS = $(shell find client -name '*.go')
BOT_SRCS = $(shell find bot -name '*.go')
COMMON_SRCS = $(shell find common -name '*.go')
PROTO_SRCS = $(shell find proto -name '*.go')

.PHONY: all build clean proto

all: build

build: $(SERVER) $(CLIENT) $(BOT)

clean:
	rm $(SERVER) $(CLIENT) $(BOT)

proto:
	protoc --go_out=plugins=grpc:proto chat.proto

$(SERVER): $(SERVER_SRCS) $(COMMON_SRCS) $(PROTO_SRCS)
	go build -o $(SERVER) github.com/etheriqa/go-grpc-chat/server

$(CLIENT): $(CLIENT_SRCS) $(COMMON_SRCS) $(PROTO_SRCS)
	go build -o $(CLIENT) github.com/etheriqa/go-grpc-chat/client

$(BOT): $(BOT_SRCS) $(COMMON_SRCS) $(PROTO_SRCS)
	go build -o $(BOT) github.com/etheriqa/go-grpc-chat/bot
