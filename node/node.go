package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"strconv"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/boltdb/bolt"
	pb "github.com/freddygv/cassandra-wannabe/pb/crud"
	"google.golang.org/grpc"
)

const (
	port   = ":8081"
	bucket = "ratings"
)

type crudServer struct {
	db *bolt.DB
}

func main() {
	db, err := bolt.Open("data/node.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatalf("failed to open: %v", err)
	}
	initializeDB(db)

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterCRUDServiceServer(s, &crudServer{db})
	s.Serve(lis)

	db.Close()
}

func initializeDB(db *bolt.DB) {
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte(bucket))
		if err != nil {
			log.Fatalf("create bucket: %v", err)
		}
		return nil
	})
}

func (s *crudServer) Read(ctx context.Context, in *pb.Key) (*pb.Record, error) {
	var record pb.Record

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))

		key := strconv.Itoa(int(in.UserID)) // TODO: Replace with murmur3

		v := b.Get([]byte(key))
		err := json.Unmarshal(v, &record)
		return err

	})

	return &record, err
}

func (s *crudServer) Upsert(ctx context.Context, in *pb.Record) (*pb.UpsertResponse, error) {
	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))

		key := strconv.Itoa(int(in.UserID)) // TODO: Replace with murmur3

		buf, err := json.Marshal(in)
		if err != nil {
			return err
		}

		err = b.Put([]byte(key), buf)
		return err
	})
	if err != nil {
		return &pb.UpsertResponse{}, status.Errorf(codes.Internal, "update: %v", err)
	}

	return &pb.UpsertResponse{}, nil
}

func (s *crudServer) Delete(ctx context.Context, in *pb.Key) (*pb.DeleteResponse, error) {
	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))

		key := strconv.Itoa(int(in.UserID)) // TODO: Replace with murmur3

		err := b.Delete([]byte(key))
		return err
	})
	if err != nil {
		return &pb.DeleteResponse{}, status.Errorf(codes.Internal, "delete: %v", err)
	}

	return &pb.DeleteResponse{}, nil
}
