package main

import (
	"errors"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/etheriqa/go-grpc-chat/proto"
)

const sessionIDLength = 16

type sessionID [sessionIDLength]byte

func alignSessionID(raw []byte) sessionID {
	var sid sessionID
	for i := 0; i < sessionIDLength && i < len(raw); i++ {
		sid[i] = raw[i]
	}
	return sid
}

type chatServer struct {
	mu   sync.RWMutex
	name map[sessionID]string
	buf  map[sessionID]chan *pb.Event
}

func newChatServer() *chatServer {
	return &chatServer{
		name: make(map[sessionID]string),
		buf:  make(map[sessionID]chan *pb.Event),
	}
}

func (cs *chatServer) withReadLock(f func()) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	f()
}

func (cs *chatServer) withWriteLock(f func()) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	f()
}

func (cs *chatServer) generateSessionID() sessionID {
	var sid sessionID
	for i := 0; i < sessionIDLength/4; i++ {
		r := rand.Uint32()
		for j := 0; j < 4 && i*4+j < sessionIDLength; j++ {
			sid[i*4+j] = byte(r)
			r <<= 8
		}
	}
	return sid
}

func (cs *chatServer) unsafeExpire(sid sessionID) {
	delete(cs.name, sid)
	delete(cs.buf, sid)
}

func (cs *chatServer) Authorize(ctx context.Context, req *pb.RequestAuthorize) (*pb.ResponseAuthorize, error) {
	if len(req.Name) == 0 {
		return nil, errors.New("name must be not empty")
	}
	if len(req.Name) > 20 {
		return nil, errors.New("name must be less than or equal 20 characters")
	}

	sid := cs.generateSessionID()
	cs.withWriteLock(func() {
		cs.name[sid] = req.Name
	})
	go func() {
		time.Sleep(5 * time.Second)
		cs.withWriteLock(func() {
			if _, ok := cs.buf[sid]; ok {
				return
			}
			cs.unsafeExpire(sid)
		})
	}()

	res := pb.ResponseAuthorize{
		SessionId: sid[:],
	}
	return &res, nil
}

func (cs *chatServer) Connect(req *pb.RequestConnect, stream pb.Chat_ConnectServer) error {
	var (
		sid  sessionID      = alignSessionID(req.SessionId)
		buf  chan *pb.Event = make(chan *pb.Event)
		err  error
		name string
	)

	cs.withWriteLock(func() {
		var ok bool
		name, ok = cs.name[sid]
		if !ok {
			err = errors.New("not authorized")
			return
		}
		if _, ok := cs.buf[sid]; ok {
			err = errors.New("already connected")
			return
		}
		cs.buf[sid] = buf
	})
	if err != nil {
		return err
	}

	go cs.withReadLock(func() {
		log.Printf("Join name=%s\n", name)
		for _, buf := range cs.buf {
			buf <- &pb.Event{
				Join: &pb.EventJoin{
					Name: name,
				},
			}
		}
	})
	defer cs.withReadLock(func() {
		log.Printf("Leave name=%s\n", name)
		for _, buf := range cs.buf {
			buf <- &pb.Event{
				Leave: &pb.EventLeave{
					Name: name,
				},
			}
		}
	})
	defer cs.withWriteLock(func() { cs.unsafeExpire(sid) })

	for {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		case event := <-buf:
			if err := stream.Send(event); err != nil {
				return err
			}
		}
	}
}

func (cs *chatServer) Say(ctx context.Context, req *pb.CommandSay) (*pb.None, error) {
	var (
		sid  sessionID = alignSessionID(req.SessionId)
		name string
		err  error
	)

	cs.withReadLock(func() {
		var ok bool
		name, ok = cs.name[sid]
		if !ok {
			err = errors.New("not authorized")
			return
		}
		if _, ok := cs.buf[sid]; !ok {
			err = errors.New("not authorized")
			return
		}
	})
	if err != nil {
		return nil, err
	}

	if len(req.Message) == 0 {
		return nil, errors.New("message must be not empty")
	}
	if len(req.Message) > 140 {
		return nil, errors.New("message must be less than or equal 140 characters")
	}

	go cs.withReadLock(func() {
		log.Printf("Log name=%s message=%s\n", name, req.Message)
		for _, buf := range cs.buf {
			buf <- &pb.Event{
				Log: &pb.EventLog{
					Name:    name,
					Message: req.Message,
				},
			}
		}
	})

	return &pb.None{}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":5000")
	if err != nil {
		log.Fatalln("net.Listen:", err)
	}
	cs := newChatServer()
	go func() {
		tick := time.Tick(time.Second)
		for {
			select {
			case <-tick:
				cs.withReadLock(func() {
					for _, buf := range cs.buf {
						buf <- &pb.Event{
							Log: &pb.EventLog{
								Name:    "system",
								Message: time.Now().String(),
							},
						}
					}
				})
			}
		}
	}()
	server := grpc.NewServer()
	pb.RegisterChatServer(server, cs)
	server.Serve(lis)
}
