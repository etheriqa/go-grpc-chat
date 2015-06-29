package main

import (
	"fmt"
	"io"
	"log"

	"google.golang.org/grpc"

	"github.com/etheriqa/go-grpc-chat/common"
	pb "github.com/etheriqa/go-grpc-chat/proto"
)

func main() {
	conn, err := grpc.Dial(":5000")
	if err != nil {
		log.Fatalln("net.Dial:", err)
	}
	defer conn.Close()
	client := pb.NewChatClient(conn)

	var name string
	for {
		fmt.Print("name> ")
		if n, err := fmt.Scanln(&name); err == io.EOF {
			return
		} else if n == 0 {
			fmt.Println("name must be not empty")
			continue
		} else if n > 20 {
			fmt.Println("name must be less than or equal 20 characters")
			continue
		}
		break
	}

	sid, err := common.Authorize(client, name)
	if err != nil {
		log.Fatalln("authorize:", err)
	}

	events, err := common.Connect(client, sid)
	if err != nil {
		log.Fatalln("connect:", err)
	}

	go func() {
		for {
			select {
			case event := <-events:
				switch {
				case event.Join != nil:
					fmt.Printf("%s has joined.\n", event.Join.Name)
				case event.Leave != nil:
					fmt.Printf("%s has left.\n", event.Leave.Name)
				case event.Log != nil:
					fmt.Printf("%s> %s\n", event.Log.Name, event.Log.Message)
				}
			}
		}
	}()

	var message string
	for {
		fmt.Print("> ")
		if n, err := fmt.Scanln(&message); err == io.EOF {
			return
		} else if n > 0 {
			err := common.Say(client, sid, message)
			if err != nil {
				log.Fatalln("say:", err)
			}
		}
	}
}
