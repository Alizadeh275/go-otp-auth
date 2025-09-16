package model

import "time"

type User struct {
    ID int64 `db:"id" json:"id"`
    Phone string `db:"phone" json:"phone"`
    RegisteredAt time.Time `db:"registered_at" json:"registered_at"`
}