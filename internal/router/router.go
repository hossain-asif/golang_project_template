package router

import (
	"go_project_structure/internal/module"
	"go_project_structure/internal/user"

	"github.com/go-chi/chi/v5"
)

type Router interface {
	Register(r chi.Router)
}

var Modules = []module.Module{
	&user.UserModule{},
	// &role.RoleModule{},
}
