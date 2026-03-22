package model

import "time"

type UserData struct {
	ID        uint32
	UserID    uint32
	DataKey   string
	DataValue []byte
	UpdatedAt time.Time
	DeletedAt time.Time
}
