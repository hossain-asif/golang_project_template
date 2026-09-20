package identity

import (
	"go_project_structure/common_pkg/proxy"
	"go_project_structure/internal/pkg/middlewares"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type router struct {
	handler *Handler
}

func NewRouter(_userController *Handler) *router {
	return &router{
		handler: _userController,
	}
}

func (ur *router) Register(r chi.Router) {
	r.Mount("/api/v1/users", ur.v1())
	r.Mount("/api/v2/users", ur.v2())
}

func (ur *router) v1() http.Handler {
	r := chi.NewRouter()

	r.Use(middlewares.RequestLoggerMiddleware)

	// Public
	r.Post("/signup", ur.handler.RegisterUser)
	r.Post("/login", ur.handler.LoginUser)

	// Protected (JWT required)
	r.Group(func(r chi.Router) {
		r.Use(middlewares.JwtAuthMiddleware)

		r.Route("/profile/{id}", func(r chi.Router) {
			r.Get("/", ur.handler.GetUserById)
			r.Delete("/", ur.handler.DeleteUser)
			r.With(middlewares.RateLimitMiddleware).
				Patch("/", ur.handler.UpdateUser)
		})

		r.Get("/profile/all", ur.handler.GetAllUsers)

	})

	// proxy
	r.Get("/fake-store/*", proxy.ProxyToService("https://fakestoreapi.com", "/api/v1/fake-store"))

	return r
}

func (ur *router) v2() http.Handler {
	r := chi.NewRouter()

	r.Use(middlewares.RequestLoggerMiddleware)

	// Public
	r.Post("/signup", ur.handler.RegisterUser)
	r.Post("/login", ur.handler.LoginUser)

	// Protected (JWT required)
	r.Group(func(r chi.Router) {
		r.Use(middlewares.JwtAuthMiddleware)
		r.Get("/profile", ur.handler.GetAllUsers)

		r.Route("/profile/{id}", func(r chi.Router) {
			r.Get("/", ur.handler.GetUserById)
			r.Delete("/", ur.handler.DeleteUser)
			r.With(middlewares.RateLimitMiddleware).
				Patch("/", ur.handler.UpdateUser)
		})
	})

	return r
}
