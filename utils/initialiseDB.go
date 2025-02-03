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

	// Create Posts table
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

	// Create Reaction table
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

	// Create Triggers
	_, err = db.Exec(`
CREATE TRIGGER IF NOT EXISTS AfterReactionInsert
AFTER INSERT ON reaction
BEGIN
    UPDATE posts SET
        likes = CASE 
            WHEN NEW.like = 1 THEN likes + 1 
            ELSE likes 
        END,
        dislikes = CASE 
            WHEN NEW.like = 0 THEN dislikes + 1 
            ELSE dislikes 
        END
    WHERE id = NEW.post_id;
END;
`)
	if err != nil {
		return nil, fmt.Errorf("failed to create triggers: %v", err)
	}

	_, err = db.Exec(`
CREATE TRIGGER IF NOT EXISTS AfterReactionUpdate
AFTER UPDATE ON reaction
BEGIN
    UPDATE posts SET
        likes = CASE 
            WHEN OLD.like = 1 THEN likes - 1
            WHEN NEW.like = 1 THEN likes + 1
            ELSE likes 
        END,
        dislikes = CASE 
            WHEN OLD.like = 0 THEN dislikes - 1
            WHEN NEW.like = 0 THEN dislikes + 1
            ELSE dislikes 
        END
    WHERE id = NEW.post_id;
END;
`)
	if err != nil {
		return nil, fmt.Errorf("failed to create triggers: %v", err)
	}

	_, err = db.Exec(`
CREATE TRIGGER IF NOT EXISTS AfterReactionDelete
AFTER DELETE ON reaction
BEGIN
    UPDATE posts SET
        likes = CASE 
            WHEN OLD.like = 1 THEN likes - 1 
            ELSE likes 
        END,
        dislikes = CASE 
            WHEN OLD.like = 0 THEN dislikes - 1 
            ELSE dislikes 
        END
    WHERE id = OLD.post_id;
END;
`)
	if err != nil {
		return nil, fmt.Errorf("failed to create triggers: %v", err)
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
	// Create Comment Reaction table
	_, err = db.Exec(`
    CREATE TABLE IF NOT EXISTS comment_reaction (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id TEXT NOT NULL,
        comment_id INTEGER NOT NULL,
        is_like INTEGER NOT NULL CHECK (is_like IN (0, 1)),
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
        FOREIGN KEY (comment_id) REFERENCES comments(id) ON DELETE CASCADE,
        UNIQUE(user_id, comment_id)
    );
    CREATE INDEX IF NOT EXISTS idx_comment_reaction_comment_id ON comment_reaction(comment_id);
    CREATE INDEX IF NOT EXISTS idx_comment_reaction_user_id ON comment_reaction(user_id);
`)
	if err != nil {
		return nil, fmt.Errorf("failed to create comment_reaction table: %v", err)
	}

	// Create triggers to update comment likes/dislikes after insert, update and delete
	_, err = db.Exec(`
CREATE TRIGGER IF NOT EXISTS AfterCommentReactionInsert
AFTER INSERT ON comment_reaction
BEGIN
    UPDATE comments SET
        likes = CASE WHEN NEW.is_like = 1 THEN likes + 1 ELSE likes END,
        dislikes = CASE WHEN NEW.is_like = 0 THEN dislikes + 1 ELSE dislikes END
    WHERE id = NEW.comment_id;
END;

`)
	if err != nil {
		return nil, fmt.Errorf("failed to create trigger AfterCommentReactionInsert: %v", err)
	}

	_, err = db.Exec(`
CREATE TRIGGER IF NOT EXISTS AfterCommentReactionUpdate
AFTER UPDATE ON comment_reaction
BEGIN
    UPDATE comments SET
        likes = CASE 
                    WHEN OLD.is_like = 1 THEN likes - 1
                    WHEN NEW.is_like = 1 THEN likes + 1
                    ELSE likes 
                END,
        dislikes = CASE 
                    WHEN OLD.is_like = 0 THEN dislikes - 1
                    WHEN NEW.is_like = 0 THEN dislikes + 1
                    ELSE dislikes 
                END
    WHERE id = NEW.comment_id;
END;

`)
	if err != nil {
		return nil, fmt.Errorf("failed to create trigger AfterCommentReactionUpdate: %v", err)
	}

	_, err = db.Exec(`
CREATE TRIGGER IF NOT EXISTS AfterCommentReactionDelete
AFTER DELETE ON comment_reaction
BEGIN
    UPDATE comments SET
        likes = CASE WHEN OLD.is_like = 1 THEN likes - 1 ELSE likes END,
        dislikes = CASE WHEN OLD.is_like = 0 THEN dislikes - 1 ELSE dislikes END
    WHERE id = OLD.comment_id;
END;

`)
	if err != nil {
		return nil, fmt.Errorf("failed to create trigger AfterCommentReactionDelete: %v", err)
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
	err = InsertDefaultCategories()
	if err != nil {
		return nil, fmt.Errorf("failed to insert default categories: %v", err)
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

	return db, nil
}

func InsertDefaultCategories() error {
	categories := []string{
		"Tech",
		"Programming",
		"Business",
		"Lifestyle",
		"Personal Development",
		"Football",
		"Politics",
		"General News",
	}

	for _, category := range categories {
		_, err := GlobalDB.Exec("INSERT OR IGNORE INTO categories (name) VALUES (?)", category)
		if err != nil {
			return fmt.Errorf("failed to insert category %s: %v", category, err)
		}
	}
	return nil
}
