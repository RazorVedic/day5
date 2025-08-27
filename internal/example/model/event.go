package model

import storage "github.com/razorpay/goutils/sqlstorage"

// Event is the generic db model for event table
type Event struct {
	// ObjectMeta is for object storage meta
	storage.ObjectMeta

	// Data is event data in bytes
	Data []byte `gorm:"not null" json:"data"`

	// Topic is the name of the topic that needs to be
	Topic string `gorm:"not null" json:"topic"`
}

// TableName for the model
func (e *Event) TableName() string {
	return "event"
}
