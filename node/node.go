package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"go.opencensus.io/exporter/prometheus"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
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
	db      *bolt.DB
	metrics Metrics
}

type Metrics struct {
	read        *stats.Int64Measure
	readMiss    *stats.Int64Measure
	readError   *stats.Int64Measure
	upsert      *stats.Int64Measure
	upsertError *stats.Int64Measure
	delete      *stats.Int64Measure
	deleteError *stats.Int64Measure
}

type Row struct {
	Payload *pb.Record
	Created time.Time
}

func main() {
	var server crudServer

	server.instrument()
	server.buildDB()
	server.serveCRUD()

	server.db.Close()
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

func (s *crudServer) buildDB() {
	// Start fresh
	err := os.RemoveAll(file)
	if err != nil {
		log.Fatalf("failed to delete db: %v", err)
	}
	s.db, err = bolt.Open(file, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}

	// Init bucket
	s.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte(bucket))
		if err != nil {
			log.Fatalf("failed to create bucket: %v", err)
		}
		return nil
	})

}

// Create metrics, define aggregations for them, then subscribe
func (s *crudServer) instrument() {
	exp, err := prometheus.NewExporter(prometheus.Options{})
	if err != nil {
		log.Fatalf("failed to create exporter: %v", err)
	}
	view.RegisterExporter(exp)

	s.metrics.read, err = stats.Int64("cw/measures/read_count", "number of keys read", "")
	if err != nil {
		log.Fatalf("failed to create metric: %v", err)
	}

	viewRead, err := view.New(
		metricName(s.metrics.read.Name()),
		s.metrics.read.Description(),
		nil,
		s.metrics.read,
		view.CountAggregation{},
	)
	if err != nil {
		log.Fatalf("failed to create view: %v", err)
	}
	if err := viewRead.Subscribe(); err != nil {
		log.Fatalf("failed to subscribe: %v", err)
	}

	s.metrics.readMiss, err = stats.Int64("cw/measures/read_miss_count", "number read misses (key not in db)", "")
	if err != nil {
		log.Fatalf("failed to create metric: %v", err)
	}
	viewReadMiss, err := view.New(
		metricName(s.metrics.readMiss.Name()),
		s.metrics.readMiss.Description(),
		nil,
		s.metrics.readMiss,
		view.CountAggregation{},
	)
	if err != nil {
		log.Fatalf("failed to create view: %v", err)
	}
	if err := viewReadMiss.Subscribe(); err != nil {
		log.Fatalf("failed to subscribe: %v", err)
	}

	s.metrics.readError, err = stats.Int64("cw/measures/read_error_count", "number of read errors", "")
	if err != nil {
		log.Fatalf("failed to create metric: %v", err)
	}
	viewReadError, err := view.New(
		metricName(s.metrics.readError.Name()),
		s.metrics.readError.Description(),
		nil,
		s.metrics.readError,
		view.CountAggregation{},
	)
	if err != nil {
		log.Fatalf("failed to create view: %v", err)
	}
	if err := viewReadError.Subscribe(); err != nil {
		log.Fatalf("failed to subscribe: %v", err)
	}

	s.metrics.upsert, err = stats.Int64("cw/measures/upsert_count", "number of keys upserted", "")
	if err != nil {
		log.Fatalf("failed to create metric: %v", err)
	}
	viewUpsert, err := view.New(
		metricName(s.metrics.upsert.Name()),
		s.metrics.upsert.Description(),
		nil,
		s.metrics.upsert,
		view.CountAggregation{},
	)
	if err != nil {
		log.Fatalf("failed to create view: %v", err)
	}
	if err := viewUpsert.Subscribe(); err != nil {
		log.Fatalf("failed to subscribe: %v", err)
	}

	s.metrics.upsertError, err = stats.Int64("cw/measures/upsert_error_count", "number of upsert errors", "")
	if err != nil {
		log.Fatalf("failed to create metric: %v", err)
	}
	viewUpsertError, err := view.New(
		metricName(s.metrics.upsertError.Name()),
		s.metrics.upsertError.Description(),
		nil,
		s.metrics.upsertError,
		view.CountAggregation{},
	)
	if err != nil {
		log.Fatalf("failed to create view: %v", err)
	}
	if err := viewUpsertError.Subscribe(); err != nil {
		log.Fatalf("failed to subscribe: %v", err)
	}

	s.metrics.delete, err = stats.Int64("cw/measures/delete_count", "number of keys deleted", "")
	if err != nil {
		log.Fatalf("failed to create metric: %v", err)
	}
	viewDelete, err := view.New(
		metricName(s.metrics.delete.Name()),
		s.metrics.delete.Description(),
		nil,
		s.metrics.delete,
		view.CountAggregation{},
	)
	if err != nil {
		log.Fatalf("failed to create view: %v", err)
	}
	if err := viewDelete.Subscribe(); err != nil {
		log.Fatalf("failed to subscribe: %v", err)
	}

	s.metrics.deleteError, err = stats.Int64("cw/measures/delete_error_count", "number of key deletion errors", "")
	if err != nil {
		log.Fatalf("failed to create metric: %v", err)
	}
	viewDeleteError, err := view.New(
		metricName(s.metrics.deleteError.Name()),
		s.metrics.deleteError.Description(),
		nil,
		s.metrics.deleteError,
		view.CountAggregation{},
	)
	if err != nil {
		log.Fatalf("failed to create view: %v", err)
	}
	if err := viewDeleteError.Subscribe(); err != nil {
		log.Fatalf("failed to subscribe: %v", err)
	}

	view.SetReportingPeriod(1 * time.Second)

}

// gRPC server
func (s *crudServer) serveCRUD() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	gServer := grpc.NewServer()
	pb.RegisterCRUDServiceServer(gServer, s)
	gServer.Serve(lis)
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

func metricName(full string) string {
	i := strings.LastIndex(full, "/")
	return string([]rune(full)[i+1:])
}
