package models

type Role string

const (
	RoleAdmin      Role = "Admin"
	RoleUser       Role = "User"
	RoleSuperAdmin Role = "SuperAdmin"
)
