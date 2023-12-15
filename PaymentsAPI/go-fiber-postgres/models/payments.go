package models

import (
	"time"

	"gorm.io/gorm"
)

type Payments struct {
	ID          uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	CreditorAcc *string    `json:"creditorAcc"`
	DebtorAcc   *string    `json:"debtorAcc"`
	Currency    *string    `json:"currency"`
	Amount      *float32   `json:"amount"`
	Date        *time.Time `json:"date"`
	IsDeleted   bool       `json:"isDeleted"`
}

func MigratePayments(db *gorm.DB) error {
	err := db.AutoMigrate(&Payments{})
	return err
}
