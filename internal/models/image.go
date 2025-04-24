package models

import "time"

type Image struct {
	ID        int
	UserID    int
	FilePath  string
	Likes     int
	CreatedAt time.Time
	Comments  []Comment
	IsOwner   bool
}
