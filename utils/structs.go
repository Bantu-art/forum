package utils

import "time"

type User struct {
	ID       string
	UserName string
	Email    string
	Password string
}

type Post struct {
	ID       string
	UserID   string
	Title    string
	Content  string
	PostTime time.Time
	Likes    int
	Dislikes int
	Comments int
}

type Comment struct {
	ID          string
	PostID      string
	UserID      string
	Content     string
	CommentTime time.Time
	Likes       int
	Dislikes    int
}

type Category struct {
	ID   string
	Name string
}

type Session struct {
	ID        string
	UserID    string
	ExpiresAt time.Time
}
