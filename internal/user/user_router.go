package user

import (
	"go_project_structure/common_pkg/proxy"
	"go_project_structure/internal/pkg/middlewares"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type UserRouter struct {
	userController *UserController
}

func NewUserRouter(_userController *UserController) *UserRouter {
	return &UserRouter{
		userController: _userController,
	}
}

func (ur *UserRouter) Register(r chi.Router) {
	r.Mount("/api/v1/users", ur.v1())
	r.Mount("/api/v2/users", ur.v2())
}

func (ur *UserRouter) v1() http.Handler {
	r := chi.NewRouter()

	r.Use(middlewares.RequestLoggerMiddleware)

	// Public
	r.Post("/signup", ur.userController.RegisterUser)
	r.Post("/login", ur.userController.LoginUser)

	// Protected (JWT required)
	r.Group(func(r chi.Router) {
		r.Use(middlewares.JwtAuthMiddleware)

		r.Route("/profile/{id}", func(r chi.Router) {
			r.Get("/", ur.userController.GetUserById)
			r.Delete("/", ur.userController.DeleteUser)
			r.With(middlewares.RateLimitMiddleware).
				Patch("/", ur.userController.UpdateUser)
		})

		r.Get("/profile/all", ur.userController.GetAllUsers)
		r.Get("/profile/export", ur.userController.ExportUsersCSV)
		r.Get("/profile/download", ur.userController.DownloadFileHandler)
		r.With(UserUploadCSVRequestValidator).
			Post("/profile/upload", ur.userController.UploadUserCSV)

		// pagination
		r.Get("/profile/offset", ur.userController.GetUsersByOffsetPagination)
		r.Get("/profile/cursor", ur.userController.GetUsersByCursorPagination)
		r.Get("/profile/seek", ur.userController.GetUsersBySeekPagination)

	})

	// proxy
	r.Get("/fake-store/*", proxy.ProxyToService("https://fakestoreapi.com", "/api/v1/fake-store"))

	return r
}

func (ur *UserRouter) v2() http.Handler {
	r := chi.NewRouter()

	r.Use(middlewares.RequestLoggerMiddleware)

	// Public
	r.Post("/signup", ur.userController.RegisterUser)
	r.Post("/login", ur.userController.LoginUser)

	// Protected (JWT required)
	r.Group(func(r chi.Router) {
		r.Use(middlewares.JwtAuthMiddleware)
		r.Get("/profile", ur.userController.GetAllUsers)
		r.Get("/profile/export", ur.userController.ExportUsersCSV)
		r.Get("/profile/download", ur.userController.DownloadFileHandler)
		r.Post("/profile/upload", ur.userController.UploadUserCSV)

		r.Route("/profile/{id}", func(r chi.Router) {
			r.Get("/", ur.userController.GetUserById)
			r.Delete("/", ur.userController.DeleteUser)
			r.With(middlewares.RateLimitMiddleware).
				Patch("/", ur.userController.UpdateUser)
		})
	})

	return r
}
