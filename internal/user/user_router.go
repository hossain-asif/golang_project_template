package user

import (
	common_middlewares "go_project_structure/common_pkg/middlewares"
	"go_project_structure/common_pkg/proxy"
	"go_project_structure/internal/middlewares"

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
	r.Use(common_middlewares.RequestLoggerMiddleware)

	r.With(middlewares.UserRegisterRequestValidator).
		Post("/signup", ur.userController.RegisterUser)
	r.Post("/login", ur.userController.LoginUser)

	r.With(common_middlewares.JwtAuthMiddleware).
		Get("/profile/{id}", ur.userController.GetUserById)

	r.With(common_middlewares.JwtAuthMiddleware).
		Get("/profile", ur.userController.GetAllUsers)

	r.With(common_middlewares.RateLimitMiddleware, middlewares.UserUpdateRequestValidator).
		Patch("/profile/{id}", ur.userController.UpdateUser)

	r.With(common_middlewares.JwtAuthMiddleware).
		Delete("/profile/{id}", ur.userController.DeleteUser)

	r.Get("/profile/export", ur.userController.ExportUsersCSV)
	r.Get("/profile/download", ur.userController.DownloadFileHandler)

	r.Post("/profile/upload", ur.userController.UploadUserCSV)

	// proxy routes
	r.Get("/fake-store/*", proxy.ProxyToService("https://fakestoreapi.com", "/fake-store"))
}
