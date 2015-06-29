package common

import (
	"io"
	"log"

	"golang.org/x/net/context"

	pb "github.com/etheriqa/go-grpc-chat/proto"
)

func Authorize(client pb.ChatClient, name string) (sid []byte, err error) {
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

func Connect(client pb.ChatClient, sid []byte) (events chan *pb.Event, err error) {
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

func Say(client pb.ChatClient, sid []byte, message string) error {
	req := pb.CommandSay{
		SessionId: sid,
		Message:   message,
	}
	_, err := client.Say(context.Background(), &req)
	return err
}
