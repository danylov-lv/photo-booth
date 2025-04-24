package models

import "time"

type User struct {
	ID                int
	Username          string
	Email             string
	Password          string
	ConfirmationToken string
	IsConfirmed       bool
	ResetToken        string
	ResetTokenExpiry  time.Time
	CreatedAt         time.Time
	NotifyOnComment   bool
}
