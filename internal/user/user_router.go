package user

import (
	"go_project_structure/common_pkg/proxy"
	"go_project_structure/internal/middlewares"
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
	r.Mount("/api/v1", ur.v1())
	r.Mount("/api/v2", ur.v2())
}

func (ur *UserRouter) v1() http.Handler {
	r := chi.NewRouter()

	r.Use(middlewares.RequestLoggerMiddleware)

	// Public
	r.With(middlewares.UserRegisterRequestValidator).
		Post("/signup", ur.userController.RegisterUser)
	r.With(middlewares.UserLoginRequestValidator).
		Post("/login", ur.userController.LoginUser)

	// Protected (JWT required)
	r.Group(func(r chi.Router) {
		r.Use(middlewares.JwtAuthMiddleware)
		r.Get("/profile/all", ur.userController.GetAllUsers)
		r.Get("/profile/export", ur.userController.ExportUsersCSV)
		r.Get("/profile/download", ur.userController.DownloadFileHandler)
		r.With(middlewares.UserUploadCSVRequestValidator).
			Post("/profile/upload", ur.userController.UploadUserCSV)

		// pagination
		r.Get("/profile/offset", ur.userController.GetUsersByOffsetPagination)
		r.Get("/profile/cursor", ur.userController.GetUsersByCursorPagination)
		r.Get("/profile/seek", ur.userController.GetUsersBySeekPagination)

		// file system
		r.Get("/user/file/{id}", ur.userController.GetUserFromFile)
		r.Post("/user/file", ur.userController.AddUserToFile)
		r.Patch("/user/file/{id}", ur.userController.UpdateUserInFile)
		r.Delete("/user/file/{id}", ur.userController.DeleteUserFromFile)

		r.Route("/profile/{id}", func(r chi.Router) {
			r.Get("/", ur.userController.GetUserById)
			r.Delete("/", ur.userController.DeleteUser)
			r.With(middlewares.RateLimitMiddleware, middlewares.UserUpdateRequestValidator).
				Patch("/", ur.userController.UpdateUser)
		})
	})

	// proxy
	r.Get("/fake-store/*", proxy.ProxyToService("https://fakestoreapi.com", "/api/v1/fake-store"))

	return r
}

func (ur *UserRouter) v2() http.Handler {
	r := chi.NewRouter()

	r.Use(middlewares.RequestLoggerMiddleware)

	// Public
	r.With(middlewares.UserRegisterRequestValidator).
		Post("/signup", ur.userController.RegisterUser)
	r.With(middlewares.UserLoginRequestValidator).
		Post("/login", ur.userController.LoginUser)

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
			r.With(middlewares.RateLimitMiddleware, middlewares.UserUpdateRequestValidator).
				Patch("/", ur.userController.UpdateUser)
		})
	})

	return r
}
