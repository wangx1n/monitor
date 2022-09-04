package main

import (
	"log"
	"net"

	"monitor/handler"
	"monitor/starter/grpc/consul"

	pb "github.com/wangx1n/monitor_pb/pb/helloworld"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

const (
	port = ":50051"
)

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &handler.V1ServerHandler{})
	grpc_health_v1.RegisterHealthServer(s, &consul.HealthImpl{})
	consul.RegisterToConsul()
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
