package postcontollers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"forum/utils"
)

func (ph *PostHandler) handleDeletePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("userID").(string)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		PostID int `json:"post_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := DeletePost(utils.GlobalDB, req.PostID); err != nil {
		log.Printf("Error deleting post: %v", err)
		http.Error(w, "Failed to delete post", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func DeletePost(db *sql.DB, postID int) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Order matters: delete child records first
	deleteQueries := []struct {
		query string
		table string
	}{
		{"DELETE FROM comment_reaction WHERE comment_id IN (SELECT id FROM comments WHERE post_id = ?)", "comment_reaction"},
		{"DELETE FROM reaction WHERE post_id = ?", "reaction"},
		{"DELETE FROM post_categories WHERE post_id = ?", "post_categories"},
		{"DELETE FROM comments WHERE post_id = ?", "comments"},
		{"DELETE FROM posts WHERE id = ?", "posts"},
	}

	for _, q := range deleteQueries {
		result, err := tx.Exec(q.query, postID)
		if err != nil {
			return fmt.Errorf("failed to delete from %s: %v", q.table, err)
		}
		rows, _ := result.RowsAffected()
		log.Printf("Deleted %d rows from %s for post %d", rows, q.table, postID)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}
