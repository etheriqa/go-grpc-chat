SERVER = server
CLIENT = client
SRCS = $(shell find . -name '*.go')

.PHONY: all build

all: build

build: $(SERVER) $(CLIENT)

$(SERVER): $(SRCS)
	go build -o $(SERVER) github.com/etheriqa/go-rpc-demo/server

$(CLIENT): $(SRCS)
	go build -o $(CLIENT) github.com/etheriqa/go-rpc-demo/client
