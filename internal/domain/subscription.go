package domain

import (
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ServiceName string
	Price       int
	UserID      uuid.UUID
	StartDate   time.Time
	EndDate     time.Time
}
