package module

import (
	"go_project_structure/internal/user"
)

func BuildModules() []Module {
	return []Module{
		user.NewUserModule(),
	}
}


// var Modules = []Module{
// 	&user.UserModule{},
// 	// &role.RoleModule{},
// }
