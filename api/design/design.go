// goagen bootstrap -d github.com/freddygv/cassandra-wannabe/api/design
package design

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
)

var _ = API("db", func() {
	Title("The Cassandra-Wannabe API")
	Description("CRUD API for a distributed database.")
	License(func() {
		Name("MIT")
		URL("https://github.com/freddygv/cassandra-wannabe/blob/master/LICENSE")

	})
	Host("localhost:8080")
	Scheme("http")
	BasePath("/db/v1")
})

// RatingPayload is the input type
var RatingPayload = Type("RatingPayload", func() {
	Attribute("movieId", Integer, func() {
		Minimum(1)
		Example(138493)
	})
	Attribute("userId", Integer, func() {
		Minimum(1)
		Example(56757)
	})
	Attribute("rating", Number, func() {
		Minimum(0.5)
		Maximum(5)
		Example(3.5)
	})
	Required("movieId", "userId", "rating") // Bad request if one of these is missing
})

// RatingMedia is the output type
var RatingMedia = MediaType("application/cassandra.wannabe.rating+json", func() {
	Description("A movie rating by a user")
	TypeName("rating")
	Reference(RatingPayload) // Some of these attributes are re-used from Payload type

	Attributes(func() {
		Attribute("movieId")
		Attribute("userId")
		Attribute("rating")
		Required("movieId", "userId", "rating") // Rendering of these results must have non-zero values
	})

	View("default", func() {
		Attribute("movieId")
		Attribute("userId")
		Attribute("rating")
	})
})

// Resource groups endpoints with common properties together.
var _ = Resource("rating", func() {
	Description("A movie rating by a user")

	// Actions are the endpoints
	Action("upsert", func() {
		Description("Adds a rating record")
		Routing(PUT("/"))
		Payload(RatingPayload)
		Response(NoContent)
		Response(InternalServerError)
	})

	Action("read", func() {
		Description("Retrieves a rating record")
		Routing(GET("/:movieId/:userId"))
		Params(func() {
			Param("movieId", Integer)
			Param("userId", Integer)
		})
		Response(OK, RatingMedia)
		Response(NotFound)
		Response(InternalServerError)
	})

	Action("delete", func() {
		Description("Delete rating record")
		Routing(DELETE("/:movieId/:userId"))
		Params(func() {
			Param("movieId", Integer)
			Param("userId", Integer)
		})
		Response(Accepted)
		Response(NotFound)
		Response(InternalServerError)
	})
})

var _ = Resource("health", func() {
	Action("health", func() {
		Description("Status of the API.")
		Routing(
			GET("/status"),
		)
		Response(OK, "text/plain")
	})
})

var _ = Resource("swagger", func() {
	Origin("*", func() {
		Methods("GET", "OPTIONS")
	})
	Files("/swagger.json", "public/swagger/swagger.json")
})
