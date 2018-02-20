package main

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "github.com/freddygv/cassandra-wannabe/pb/crud"
	"google.golang.org/grpc"
)

const port = ":8080"

type crudServer struct{}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterCRUDServiceServer(s, &crudServer{})
	s.Serve(lis)
}

func (s *crudServer) Read(ctx context.Context, in *pb.Key) (*pb.Record, error) {
	fmt.Println(in)
	return &pb.Record{MovieID: in.MovieID, UserID: in.UserID, Rating: 5}, nil
}

func (s *crudServer) Upsert(ctx context.Context, in *pb.Record) (*pb.UpsertResponse, error) {
	fmt.Println(in)
	return &pb.UpsertResponse{}, nil
}

func (s *crudServer) Delete(ctx context.Context, in *pb.Key) (*pb.DeleteResponse, error) {
	fmt.Println(in)
	return &pb.DeleteResponse{}, nil
}
