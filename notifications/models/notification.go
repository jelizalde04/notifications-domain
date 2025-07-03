package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Notification struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ActorID     uuid.UUID `gorm:"type:uuid;column:actorId"`
	RecipientID uuid.UUID `gorm:"type:uuid;column:recipientId"`
	Type        string    `gorm:"type:string"`
	Content     string    `gorm:"type:text"`
	Read        bool      `gorm:"type:boolean;default:false"`
	Timestamp   time.Time `gorm:"type:timestamp"`
}

func (Notification) TableName() string {
	return "Notifications"
}

// BeforeCreate hook para GORM
func (n *Notification) BeforeCreate(tx *gorm.DB) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	if n.Timestamp.IsZero() {
		n.Timestamp = time.Now()
	}
	return nil
}
