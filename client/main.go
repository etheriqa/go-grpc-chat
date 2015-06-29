package main

import (
	"fmt"
	"io"
	"log"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/etheriqa/go-grpc-chat/proto"
)

func authorize(client pb.ChatClient, name string) (sid []byte, err error) {
	req := pb.RequestAuthorize{
		Name: name,
	}
	res, err := client.Authorize(context.Background(), &req)
	if err != nil {
		return
	}
	sid = res.SessionId
	return
}

func connect(client pb.ChatClient, sid []byte) (events chan *pb.Event, err error) {
	req := pb.RequestConnect{
		SessionId: sid,
	}
	stream, err := client.Connect(context.Background(), &req)
	if err != nil {
		return
	}
	events = make(chan *pb.Event, 1000)
	go func() {
		defer func() { close(events) }()
		for {
			event, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				log.Fatalln("stream.Recv", err)
			}
			events <- event
		}
	}()
	return
}

func say(client pb.ChatClient, sid []byte, message string) error {
	req := pb.CommandSay{
		SessionId: sid,
		Message:   message,
	}
	_, err := client.Say(context.Background(), &req)
	return err
}

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

	sid, err := authorize(client, name)
	if err != nil {
		log.Fatalln("authorize:", err)
	}

	events, err := connect(client, sid)
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
			err := say(client, sid, message)
			if err != nil {
				log.Fatalln("say:", err)
			}
		}
	}
}
