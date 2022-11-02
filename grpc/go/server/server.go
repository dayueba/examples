package main

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/protobuf/types/known/fieldmaskpb"


	pb "grpcexample/pb"
)

const (
	addr = ":50051"
)

var _ pb.HelloServiceServer = (*server)(nil)
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

func (s *server) Foo(ctx context.Context, req *pb.FooRequest) (*pb.FooResponse, error) {
	// return nil, errors.New("oops")
	return nil, status.Error(codes.NotFound, "some description")
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
