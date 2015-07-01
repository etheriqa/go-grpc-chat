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
var m = flag.Float64("m", 1, "message per second")

func bot() func() {
	conn, err := grpc.Dial(":5000")
	if err != nil {
		log.Fatalln("net.Dial:", err)
	}
	client := pb.NewChatClient(conn)

	sid, err := common.Authorize(client, "bot")
	if err != nil {
		log.Fatalln("authorize:", err)
	}

	events, err := common.Connect(client, sid)
	if err != nil {
		log.Fatalln("connect:", err)
	}

	return func() {
		defer conn.Close()
		interval := float64(time.Second) / *m
		time.Sleep(time.Duration(interval * rand.Float64()))
		tick := time.Tick(time.Duration(interval))
		for {
			select {
			case <-tick:
				err := common.Say(client, sid, "foobarbazfoobarbaz")
				if err != nil {
					log.Fatalln("say:", err)
				}
			case <-events:
			}
		}
	}
}

func main() {
	flag.Parse()

	fs := make([]func(), *n)
	for i := 0; i < *n; i++ {
		fs[i] = bot()
	}
	for i := 0; i < *n; i++ {
		go fs[i]()
	}
	for {
		time.Sleep(time.Second)
	}
}
