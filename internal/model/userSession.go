package model

import (
	"time"
)

type UserSession struct {
	Base
	ID                 string    `json:"id"`
	ExpiredAccessDate  time.Time `json:"title"`
	ExpiredRefreshDate time.Time `json:"description"`
	UserID             string    `json:"user_id"`
}
