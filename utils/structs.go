package utils

import (
	"database/sql"
	"time"
)

type User struct {
	ID         string
	UserName   string
	Email      string
	Password   string
	ProfilePic sql.NullString
}

type Post struct {
    ID         int
    UserID     string
    Title      string
    Content    string
    ImagePath  string
    PostTime   string
    Likes      int
    Dislikes   int
    Comments   int
    Username   string
    ProfilePic sql.NullString
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
	UserID    string
	ExpiresAt time.Time
}

type PageData struct {
	IsLoggedIn bool
	Posts      []Post
}
