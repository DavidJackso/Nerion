package entity

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

type User struct {
	ID           int64
	Name         string
	Email        string
	Role         Role
	PasswordHash string
}
