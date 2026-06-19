package domain

import (
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID          uuid.UUID
	ServiceName string
	Price       int
	UserID      uuid.UUID
	StartDate   time.Time
	EndDate     *time.Time
}

type CreateDTO struct {
	ServiceName string
	Price       int
	UserID      uuid.UUID
	StartDate   time.Time
	EndDate     *time.Time
}

type UpdateDTO struct {
	ServiceName string
	Price       int
	UserID      uuid.UUID
	StartDate   time.Time
	EndDate     *time.Time
}
