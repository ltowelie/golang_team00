package db

import (
	"gorm.io/gorm"
	"log"
)

type Anomaly struct {
	gorm.Model
	SessionID string
	Frequency float64
	Timestamp int64
}

func (a *Anomaly) CreateRecord(db gorm.DB) {
	db.Create(a)
}

func Migrate(db gorm.DB) {
	err := db.AutoMigrate(&Anomaly{})
	if err != nil {
		log.Fatalf("Error while automigrating schema: %s", err)
	}
}
