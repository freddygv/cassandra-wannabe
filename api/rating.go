package main

import (
	"github.com/freddygv/cassandra-wannabe/app"
	"github.com/goadesign/goa"
)

// RatingController implements the rating resource.
type RatingController struct {
	*goa.Controller
}

// NewRatingController creates a rating controller.
func NewRatingController(service *goa.Service) *RatingController {
	return &RatingController{Controller: service.NewController("RatingController")}
}

// Delete runs the delete action.
func (c *RatingController) Delete(ctx *app.DeleteRatingContext) error {
	// RatingController_Delete: start_implement

	// Put your logic here

	return nil
	// RatingController_Delete: end_implement
}

// Read runs the read action.
func (c *RatingController) Read(ctx *app.ReadRatingContext) error {
	// RatingController_Read: start_implement

	// Put your logic here

	res := &app.Rating{}
	return ctx.OK(res)
	// RatingController_Read: end_implement
}

// Upsert runs the upsert action.
func (c *RatingController) Upsert(ctx *app.UpsertRatingContext) error {
	// RatingController_Upsert: start_implement

	// Put your logic here

	return nil
	// RatingController_Upsert: end_implement
}
