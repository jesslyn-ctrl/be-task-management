package model

import (
	"github.com/google/uuid"
	"time"
)

type Base struct {
	CreatedAt  time.Time  `json:"created_at"`
	ModifiedAt time.Time  `json:"modified_at"`
	CreatedBy  *uuid.UUID `json:"created_by"`
	ModifiedBy *uuid.UUID `json:"modified_by"`
}
