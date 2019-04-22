package models

const (
	ROLE_STUDENT string = "student"
	ROLE_TEACHER string = "teacher"
	ROLE_ADMIN   string = "admin"
)

// Role 用户角色
type Role struct {
	Name string `bson:"name" json:"name"`
}
