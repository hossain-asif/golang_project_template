package example

import (
	"go_project_structure/internal/pkg/middlewares"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type router struct {
	handler *Handler
}

func NewRouter(_handler *Handler) *router {
	return &router{
		handler: _handler,
	}
}

func (ur *router) Register(r chi.Router) {
	r.Mount("/api/v1/example", ur.v1())
	r.Mount("/api/v2/example", ur.v2())
}

func (ur *router) v1() http.Handler {
	r := chi.NewRouter()

	r.Use(middlewares.RequestLoggerMiddleware)

	r.Group(func(r chi.Router) {
		r.Use(middlewares.JwtAuthMiddleware)

		r.Get("/example", ur.handler.Get)

	})

	return r
}

func (ur *router) v2() http.Handler {
	r := chi.NewRouter()

	r.Use(middlewares.RequestLoggerMiddleware)

	// Protected (JWT required)
	r.Group(func(r chi.Router) {
		r.Use(middlewares.JwtAuthMiddleware)
		r.Get("/example", ur.handler.Get)
	})

	return r
}
