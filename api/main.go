//go:generate goagen bootstrap -d github.com/freddygv/cassandra-wannabe/api/design

package main

import (
	"container/ring"

	"github.com/freddygv/cassandra-wannabe/app"
	"github.com/goadesign/goa"
	"github.com/goadesign/goa/middleware"
)

const ()

var addresses = []string{"localhost:8080"}
var addressRing *ring.Ring

func main() {
	addressRing = ring.New(len(addresses))
	// Initialize circular buffer
	for _, v := range addresses {
		addressRing.Value = v
	}

	// Create service
	service := goa.New("db")

	// Mount middleware
	service.Use(middleware.RequestID())
	service.Use(middleware.LogRequest(true))
	service.Use(middleware.ErrorHandler(service, true))
	service.Use(middleware.Recover())

	// Mount "health" controller
	c := NewHealthController(service)
	app.MountHealthController(service, c)
	// Mount "rating" controller
	c2 := NewRatingController(service)
	app.MountRatingController(service, c2)
	// Mount "swagger" controller
	c3 := NewSwaggerController(service)
	app.MountSwaggerController(service, c3)

	// Start service
	if err := service.ListenAndServe(":8080"); err != nil {
		service.LogError("startup", "err", err)
	}

}
