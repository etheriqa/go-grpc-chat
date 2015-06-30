package main

import (
	"flag"
	"log"
	"math/rand"
	"time"

	"google.golang.org/grpc"

	"github.com/etheriqa/go-grpc-chat/common"
	pb "github.com/etheriqa/go-grpc-chat/proto"
)

var n = flag.Int("n", 1, "number of client")
var m = flag.Int("m", 1, "query per second")

func bot() {
	conn, err := grpc.Dial(":5000")
	if err != nil {
		log.Fatalln("net.Dial:", err)
	}
	defer conn.Close()
	client := pb.NewChatClient(conn)

	sid, err := common.Authorize(client, "bot")
	if err != nil {
		log.Fatalln("authorize:", err)
	}

	events, err := common.Connect(client, sid)
	if err != nil {
		log.Fatalln("connect:", err)
	}

	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
	tick := time.Tick(time.Second / time.Duration(*m))
	for {
		select {
		case <-tick:
			err := common.Say(client, sid, "hi")
			if err != nil {
				log.Fatalln("say:", err)
			}
		case <-events:
		}
	}
}

func main() {
	flag.Parse()

	for i := 0; i < *n; i++ {
		go bot()
	}
	for {
		time.Sleep(time.Second)
	}
}
