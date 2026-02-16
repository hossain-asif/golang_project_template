package router

import (
	"go_project_structure/internal/middlewares"
	"go_project_structure/internal/user"
	"go_project_structure/utils"

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
	ur := user.NewUserRepository(db)
	us := user.NewUserService(ur)
	uc := user.NewUserController(us)
	uRouter := NewUserRouter(uc)
	return uRouter
}

func (ur *UserRouter) Register(r chi.Router) {
	r.Use(middlewares.RequestLoggerMiddleware)
	r.With(user.UserRegisterRequestValidator).Post("/signup", ur.userController.RegisterUser)
	r.Post("/login", ur.userController.LoginUser)
	r.With(middlewares.JwtAuthMiddleware).Get("/profile/{id}", ur.userController.GetUserById)
	r.With(middlewares.JwtAuthMiddleware).Get("/profile", ur.userController.GetAllUsers)
	r.With(middlewares.RateLimitMiddleware, user.UserUpdateRequestValidator).Patch("/profile/{id}", ur.userController.UpdateUser)
	r.With(middlewares.JwtAuthMiddleware).Delete("/profile/{id}", ur.userController.DeleteUser)

	// proxy routes
	r.Get("/fake-store/*", utils.ProxyToService("https://fakestoreapi.com", "/fake-store"))
}
