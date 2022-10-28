package main

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	pb "grpcexample/pb"
)

const (
	addr = ":50051"
)

type server struct {
	pb.UnimplementedHelloServiceServer
}

func (s *server) Hello(ctx context.Context, req *pb.Request) (*pb.Response, error) {
	log.Printf("recv: %v", req.FieldMask.Paths)
	res := pb.Response{
		FieldMask: &fieldmaskpb.FieldMask{
			Paths: []string{"hello", "world"},
		}}

	return &res, nil
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
