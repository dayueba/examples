package main

import (
	"context"
	"log"

	pb "grpcexample/pb"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

const (
	addr = ":50051"
)

func main() {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewHelloServiceClient(conn)

	ctx := context.Background()
	// ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	// defer cancel()
	r, err := c.Hello(ctx, &pb.Request{FieldMask: &fieldmaskpb.FieldMask{
		Paths: []string{"hello", "world"},
	}})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %v", r.GetFieldMask())

	_, err = c.Foo(ctx, &pb.FooRequest{})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			log.Println(err)
			return
		}

		switch st.Code() {
		case codes.InvalidArgument:
			for _, d := range st.Details() {
				switch info := d.(type) {
				case *errdetails.ErrorInfo:
					//info.Reason
					//info.Metadata
					log.Printf("Request Field Invalid: %s", info)
				default:
					log.Printf("Unexpected error type: %s", info)
				}
			}
		default:
			log.Printf("Unhandled error : %s ", st.String())
		}
	}
}
