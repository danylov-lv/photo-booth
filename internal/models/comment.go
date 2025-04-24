package models

import "time"

type Comment struct {
	ID        int
	ImageID   int
	UserID    int
	Username  string
	Content   string
	CreatedAt time.Time
}
