package handler

import (
	"context"
	pb "github.com/wangx1n/monitor_pb/pb/helloworld"
	"log"
)

type V1ServerHandler struct {
}

// SayHello implements helloworld.GreeterServer
func (s *V1ServerHandler) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", in.Name)
	return &pb.HelloReply{Message: "Hello " + in.Name}, nil
}
