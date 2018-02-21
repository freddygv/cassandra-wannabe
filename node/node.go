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
	db *bolt.DB
}

type Row struct {
	Payload *pb.Record
	Created time.Time
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

func main() {
	// Start with empty DB
	if err := os.RemoveAll(file); err != nil {
		log.Fatalf("failed to delete db: %v", err)
	}
	db, err := bolt.Open(file, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}

	// Init bucket
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte(bucket))
		if err != nil {
			log.Fatalf("failed to create bucket: %v", err)
		}
		return nil
	})

	// Prometheus exporter
	exp, err := prometheus.NewExporter(prometheus.Options{})
	if err != nil {
		log.Fatalf("failed to create exporter: %v", err)
	}
	view.RegisterExporter(exp)

	// Metrics
	var m Metrics
	err = createAndSubscribe(&m)
	if err != nil {
		log.Fatalf("failed to create metrics: %v", err)
	}

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

func createAndSubscribe(m *Metrics) error {
	var err error

	m.read, err = stats.Int64("cw/measures/read_count", "number of keys read", "")
	if err != nil {
		return err
	}

	viewRead, err := view.New(
		metricName(m.read.Name()),
		m.read.Description(),
		nil,
		m.read,
		view.CountAggregation{},
	)
	if err != nil {
		return err
	}
	if err := viewRead.Subscribe(); err != nil {
		return err
	}

	m.readMiss, err = stats.Int64("cw/measures/read_miss_count", "number read misses (key not in db)", "")
	if err != nil {
		return err
	}
	viewReadMiss, err := view.New(
		metricName(m.readMiss.Name()),
		m.readMiss.Description(),
		nil,
		m.readMiss,
		view.CountAggregation{},
	)
	if err != nil {
		return err
	}
	if err := viewReadMiss.Subscribe(); err != nil {
		return err
	}

	m.readError, err = stats.Int64("cw/measures/read_error_count", "number of read errors", "")
	if err != nil {
		return err
	}
	viewReadError, err := view.New(
		metricName(m.readError.Name()),
		m.readError.Description(),
		nil,
		m.readError,
		view.CountAggregation{},
	)
	if err != nil {
		return err
	}
	if err := viewReadError.Subscribe(); err != nil {
		return err
	}

	m.upsert, err = stats.Int64("cw/measures/upsert_count", "number of keys upserted", "")
	if err != nil {
		return err
	}
	viewUpsert, err := view.New(
		metricName(m.upsert.Name()),
		m.upsert.Description(),
		nil,
		m.upsert,
		view.CountAggregation{},
	)
	if err != nil {
		return err
	}
	if err := viewUpsert.Subscribe(); err != nil {
		return err
	}

	m.upsertError, err = stats.Int64("cw/measures/upsert_error_count", "number of upsert errors", "")
	if err != nil {
		return err
	}
	viewUpsertError, err := view.New(
		metricName(m.upsertError.Name()),
		m.upsertError.Description(),
		nil,
		m.upsertError,
		view.CountAggregation{},
	)
	if err != nil {
		return err
	}
	if err := viewUpsertError.Subscribe(); err != nil {
		return err
	}

	m.delete, err = stats.Int64("cw/measures/delete_count", "number of keys deleted", "")
	if err != nil {
		return err
	}
	viewDelete, err := view.New(
		metricName(m.delete.Name()),
		m.delete.Description(),
		nil,
		m.delete,
		view.CountAggregation{},
	)
	if err != nil {
		return err
	}
	if err := viewDelete.Subscribe(); err != nil {
		return err
	}

	m.deleteError, err = stats.Int64("cw/measures/delete_error_count", "number of key deletion errors", "")
	if err != nil {
		return err
	}
	viewDeleteError, err := view.New(
		metricName(m.deleteError.Name()),
		m.deleteError.Description(),
		nil,
		m.deleteError,
		view.CountAggregation{},
	)
	if err != nil {
		return err
	}
	if err := viewDeleteError.Subscribe(); err != nil {
		return err
	}

	view.SetReportingPeriod(1 * time.Second)

	return nil
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

func metricName(full string) string {
	i := strings.LastIndex(full, "/")
	return string([]rune(full)[i+1:])
}
