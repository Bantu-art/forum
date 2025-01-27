package utils

import "time"

type User struct {
	ID       string
	UserName string
	Email    string
	Password string
}

type Post struct {
	ID       int
	UserID   string
	Title    string
	Content  string
	PostTime time.Time
	Likes    int
	Dislikes int
	Comments int
}

type Comment struct {
	ID          int
	PostID      int
	UserID      string
	Content     string
	CommentTime time.Time
	Likes       int
	Dislikes    int
}

type Category struct {
	ID   int
	Name string
}

type Session struct {
	ID        string
	UserID    int
	ExpiresAt time.Time
}
