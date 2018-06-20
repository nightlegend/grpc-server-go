package test

import (
	"context"
	"fmt"
	"time"

	pb "github.com/nightlegend/grpc-server-go/proto"
)

// Server is define a struct for grpc
type Server struct {
	ID int
}

// GetName is get name from grpc server
func (s *Server) GetName(ctx context.Context, req *pb.Request) (*pb.Response, error) {
	if req.Id == "1" {
		return &pb.Response{Name: "Peter"}, nil
	}
	fmt.Printf("%v: Receive is %s\n", time.Now(), req.Id)
	return &pb.Response{Name: "David Guo"}, nil
}

// Echo is echo somthing.
func (s *Server) Echo(ctx context.Context, req *pb.StringMessage) (*pb.StringMessage, error) {
	return &pb.StringMessage{Value: "get value from grpc Server"}, nil
}

// GetInfo is get grpc server info.
func (s *Server) GetInfo(ctx context.Context, req *pb.Request) (*pb.Response, error) {
	fmt.Printf("the request from: %d\n", s.ID)
	return &pb.Response{Name: "Grpc-server version 1.0"}, nil
}
