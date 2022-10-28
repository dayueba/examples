package main

import (
	"net"
	"context"
	"log"

	"google.golang.org/grpc"
	pb "grpcexample/pb"
)

const (
	addr = ":50051"
)

type server struct {
	pb.UnimplementedHelloServiceServer
}

func (s *server) Hello(ctx context.Context, req *pb.String) (*pb.String, error) {
	log.Printf("recv: %v", req)
	return &pb.String{
		Value: req.Value,
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	s := grpc.NewServer()
	pb.RegisterHelloServiceServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}