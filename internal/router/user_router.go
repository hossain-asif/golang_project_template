package router

import (
	"go_project_structure/internal/db/repositories"
	"go_project_structure/internal/middlewares"
	"go_project_structure/internal/user"
	"go_project_structure/common_pkg/proxy"
	common_middlewares "go_project_structure/common_pkg/middlewares"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type UserRouter struct {
	userController *user.UserController
}

func NewUserRouter(_userController *user.UserController) *UserRouter {
	return &UserRouter{
		userController: _userController,
	}
}

func RegisterRoutes(db *gorm.DB, router chi.Router) *UserRouter {
	ur := repositories.NewUserRepository(db)
	us := user.NewUserService(ur)
	uc := user.NewUserController(us)
	uRouter := NewUserRouter(uc)
	return uRouter
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
