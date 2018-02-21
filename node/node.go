package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/boltdb/bolt"
	pb "github.com/freddygv/cassandra-wannabe/pb/crud"
	"github.com/spaolacci/murmur3"
	"google.golang.org/grpc"
)

const (
	port   = ":8081"
	bucket = "ratings"
	file   = "data/node.db"
)

type crudServer struct {
	db *bolt.DB
}

type Row struct {
	Payload *pb.Record
	Created time.Time
}

func main() {
	// Start with empty DB
	if err := os.RemoveAll(file); err != nil {
		log.Fatalf("failed to delete: %v", err)
	}
	db, err := bolt.Open(file, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatalf("failed to open: %v", err)
	}

	// Init bucket
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte(bucket))
		if err != nil {
			log.Fatalf("create bucket: %v", err)
		}
		return nil
	})

	// gRPC server
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterCRUDServiceServer(s, &crudServer{db})
	s.Serve(lis)

	db.Close()
}

func (s *crudServer) Read(ctx context.Context, in *pb.Key) (*pb.Record, error) {
	var r Row

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		k := partition(in.UserID, in.MovieID)
		v := b.Get([]byte(k))

		if v == nil {
			return status.Error(codes.NotFound, "key not found")
		}
		err := json.Unmarshal(v, &r)
		return err
	})
	if err != nil {
		return &pb.Record{}, status.Errorf(codes.Internal, "read: %v", err)
	}

	return r.Payload, nil
}

func (s *crudServer) Upsert(ctx context.Context, in *pb.Record) (*pb.UpsertResponse, error) {
	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))

		buf, err := json.Marshal(Row{Payload: in, Created: time.Now()})
		if err != nil {
			return err
		}

		k := partition(in.UserID, in.MovieID)
		err = b.Put([]byte(k), buf)
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
		k := partition(in.UserID, in.MovieID)
		err := b.Delete([]byte(k))
		return err
	})
	if err != nil {
		return &pb.DeleteResponse{}, status.Errorf(codes.Internal, "delete: %v", err)
	}

	return &pb.DeleteResponse{}, nil
}

// Combine and hash UserID and MovieID
func partition(UserID, MovieID int32) string {
	u := formatInt32(UserID)
	m := formatInt32(MovieID)
	checksum := murmur3.Sum64([]byte(u + ":" + m))

	return strconv.FormatUint(checksum, 16)
}

func formatInt32(num int32) string {
	return strconv.FormatInt(int64(num), 10)
}
