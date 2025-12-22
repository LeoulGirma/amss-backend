package domain

type Role string

const (
	RoleAdmin       Role = "admin"
	RoleTenantAdmin Role = "tenant_admin"
	RoleScheduler   Role = "scheduler"
	RoleMechanic    Role = "mechanic"
	RoleAuditor     Role = "auditor"
)

func (r Role) IsAdmin() bool {
	return r == RoleAdmin
}
