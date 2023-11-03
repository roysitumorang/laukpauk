package model

const (
	RoleAnonymous int64 = iota
	RoleSuperAdmin
	RoleAdmin
	RoleSeller
	RoleBuyer
)

type (
	Role struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}
)
