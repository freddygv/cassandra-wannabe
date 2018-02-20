package main

import (
	"context"

	"github.com/freddygv/cassandra-wannabe/app"
	pb "github.com/freddygv/cassandra-wannabe/pb/crud"
	"github.com/goadesign/goa"
	"google.golang.org/grpc"
)

// RatingController implements the rating resource.
type RatingController struct {
	*goa.Controller
}

func getAddress() string {
	addressRing.Next()

	if str, ok := addressRing.Value.(string); ok {
		return str
	}

	// TODO: Retry policy looping over all entries in the ring, then panic
	panic("No valid address to connect to")
}

// TODO: Retry policy instead of straight to 5xx resp
func dialCRUD(address string) (*grpc.ClientConn, pb.CRUDServiceClient, error) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, nil, err
	}

	return conn, pb.NewCRUDServiceClient(conn), nil
}

// NewRatingController creates a rating controller.
func NewRatingController(service *goa.Service) *RatingController {
	return &RatingController{Controller: service.NewController("RatingController")}
}

// Delete runs the delete action.
func (c *RatingController) Delete(ctx *app.DeleteRatingContext) error {
	conn, client, err := dialCRUD(getAddress())
	if err != nil {
		return ctx.InternalServerError()
	}

	// TODO: Handle NotFound case, probably as gRPC error
	if _, err := client.Delete(context.Background(),
		&pb.Key{MovieID: int32(ctx.MovieID), UserID: int32(ctx.UserID)}); err != nil {
		return ctx.InternalServerError()
	}

	conn.Close()
	return ctx.Accepted()
}

// Read runs the read action.
func (c *RatingController) Read(ctx *app.ReadRatingContext) error {
	conn, client, err := dialCRUD(getAddress())
	if err != nil {
		return ctx.InternalServerError()
	}

	r, err := client.Read(context.Background(),
		&pb.Key{MovieID: int32(ctx.MovieID), UserID: int32(ctx.UserID)})

	// TODO: Handle NotFound case, probably as gRPC error
	if err != nil {
		return ctx.InternalServerError()
	}

	res := &app.Rating{MovieID: int(r.MovieID),
		UserID: int(r.UserID),
		Rating: float64(r.Rating)}

	conn.Close()
	return ctx.OK(res)
}

// Upsert runs the upsert action.
func (c *RatingController) Upsert(ctx *app.UpsertRatingContext) error {
	conn, client, err := dialCRUD(getAddress())
	if err != nil {
		return ctx.InternalServerError()
	}

	_, err = client.Upsert(context.Background(),
		&pb.Record{MovieID: int32(ctx.Payload.MovieID),
			UserID: int32(ctx.Payload.UserID),
			Rating: float32(ctx.Payload.Rating)})

	if err != nil {
		return ctx.InternalServerError()
	}

	conn.Close()
	return ctx.NoContent()
}
