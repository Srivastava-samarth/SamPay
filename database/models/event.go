package database

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type Event struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey"`
	EventID     uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex"`
	EventType   string         `gorm:"type:varchar(100);not null"`
	Direction   string         `gorm:"type:varchar(20);not null"`
	Payload     datatypes.JSON `gorm:"type:jsonb;not null"`

	Status      string     `gorm:"type:varchar(20);not null;default:pending"`

	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

func (Event) TableName() string {
	return "events"
}