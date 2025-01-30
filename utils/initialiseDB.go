package utils

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

var GlobalDB *sql.DB

func InitialiseDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		return nil, err
	}

	GlobalDB = db

	// Create Users table
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            id TEXT PRIMARY KEY NOT NULL,
            email TEXT UNIQUE,
            username TEXT UNIQUE,
            password TEXT,
			profile_pic TEXT
        );
		CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
    	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
    `)
	if err != nil {
		return nil, fmt.Errorf("failed to create users table: %v", err)
	}

	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS posts (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT,
        title TEXT NOT NULL,
        content TEXT NOT NULL,
		imagepath TEXT,
        post_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        likes INTEGER DEFAULT 0,
		dislikes INTEGER DEFAULT 0,
		comments INTEGER DEFAULT 0,
				userreaction INTEGER,

        FOREIGN KEY (user_id) REFERENCES users(id)
    );
    CREATE INDEX IF NOT EXISTS idx_posts_user_id ON posts(user_id);
    CREATE INDEX IF NOT EXISTS idx_posts_post_at ON posts(post_at);
    `)
	if err != nil {
		return nil, fmt.Errorf("failed to create posts table: %v", err)
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS comments (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        post_id INTEGER,
        user_id TEXT,
        content TEXT NOT NULL,
        comment_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        likes INTEGER DEFAULT 0,
        dislikes INTEGER DEFAULT 0,
        FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
        FOREIGN KEY (user_id) REFERENCES users(id)
    );
    CREATE INDEX IF NOT EXISTS idx_comments_post_id ON comments(post_id);
    CREATE INDEX IF NOT EXISTS idx_comments_user_id ON comments(user_id);
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to create comments table: %v", err)
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS categories (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT UNIQUE NOT NULL
    );
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to create categories table: %v", err)
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS post_categories (
        post_id INTEGER,
        category_id INTEGER,
        PRIMARY KEY (post_id, category_id),
        FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
        FOREIGN KEY (category_id) REFERENCES categories(id)
    );
    CREATE INDEX IF NOT EXISTS idx_post_categories_category_id ON post_categories(category_id);
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to create post_categories table: %v", err)
	}
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS sessions (
        id TEXT PRIMARY KEY,
        user_id TEXT,
        expires_at DATETIME,
        FOREIGN KEY (user_id) REFERENCES users(id)
    );
    CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
    CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to create sessions table: %v", err)
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS reaction (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id TEXT NOT NULL,
        post_id INTEGER NOT NULL,
        like INTEGER NOT NULL CHECK (like IN (0, 1)), -- 1 for like, 0 for dislike
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
        FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
        UNIQUE(user_id, post_id) -- Prevent multiple reactions from same user
    );
    CREATE INDEX IF NOT EXISTS idx_reaction_post_id ON reaction(post_id);
    CREATE INDEX IF NOT EXISTS idx_reaction_user_id ON reaction(user_id);
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to create reaction table: %v", err)
	}

	// Add columns to posts table for reaction counts
	// _, err = db.Exec(`
    // ALTER TABLE posts ADD COLUMN IF NOT EXISTS likes INTEGER DEFAULT 0;
    // ALTER TABLE posts ADD COLUMN IF NOT EXISTS dislikes INTEGER DEFAULT 0;
	// `)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to add reaction columns to posts: %v", err)
	// }

	return db, nil
}
