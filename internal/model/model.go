package model

import "time"

type User struct {
	ID        int64
	Email     string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Post struct {
	ID        int64
	UserID    int64
	Title     string
	Body      string
	CreatedAt time.Time
}

type Comment struct {
	ID        int64
	PostID    int64
	UserID    int64
	Body      string
	CreatedAt time.Time
}

type Tag struct {
	ID   int64
	Name string
}
