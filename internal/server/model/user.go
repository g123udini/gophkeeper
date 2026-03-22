package model

type User struct {
	ID                 uint32
	Login              string
	PasswordHash       string
	MasterPasswordHash string
}
