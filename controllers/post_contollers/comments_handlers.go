package postcontollers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"forum/utils"
)

func (ph *PostHandler) handleComment(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(r.FormValue("post_id"))
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	content := r.FormValue("content")
	if content == "" {
		http.Error(w, "Comment cannot be empty", http.StatusBadRequest)
		return
	}

	// Insert comment
	_, err = utils.GlobalDB.Exec(`
        INSERT INTO comments (post_id, user_id, content) 
        VALUES (?, ?, ?)`,
		postID, userID, content,
	)
	if err != nil {
		log.Printf("Error creating comment: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Update post comment count
	_, err = utils.GlobalDB.Exec(`
        UPDATE posts SET comments = comments + 1 
        WHERE id = ?`, postID)
	if err != nil {
		log.Printf("Error updating comment count: %v", err)
	}

	http.Redirect(w, r, fmt.Sprintf("/?id=%d", postID), http.StatusSeeOther)
}

// handleCommentReactions processes a user's like/dislike on a comment.
func (ph *PostHandler) handleCommentReactions(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	if userID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	var req struct {
		CommentID int `json:"comment_id"`
		Like      int `json:"like"` // 1 for like, 0 for dislike
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	if req.Like != 0 && req.Like != 1 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid reaction type"})
		return
	}

	// Check if the user already reacted to this comment by querying comment_reaction
	var existingIsLike int
	err := utils.GlobalDB.QueryRow("SELECT is_like FROM comment_reaction WHERE user_id = ? AND comment_id = ?", userID, req.CommentID).Scan(&existingIsLike)
	if err != nil && err != sql.ErrNoRows {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Database error (select)"})
		return
	}

	if err == sql.ErrNoRows {
		// No reaction exists—insert a new reaction.
		_, err = utils.GlobalDB.Exec("INSERT INTO comment_reaction (user_id, comment_id, is_like) VALUES (?, ?, ?)", userID, req.CommentID, req.Like)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Database error (insert)"})
			return
		}
	} else {
		if existingIsLike == req.Like {
			// Same reaction exists; remove it.
			_, err = utils.GlobalDB.Exec("DELETE FROM comment_reaction WHERE user_id = ? AND comment_id = ?", userID, req.CommentID)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": "Database error (delete)"})
				return
			}
		} else {
			// Reaction exists but is different; update it.
			_, err = utils.GlobalDB.Exec("UPDATE comment_reaction SET is_like = ? WHERE user_id = ? AND comment_id = ?", req.Like, userID, req.CommentID)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": "Database error (update)"})
				return
			}
		}
	}

	// Fetch updated like and dislike counts for the comment.
	var likes, dislikes int
	err = utils.GlobalDB.QueryRow("SELECT likes, dislikes FROM comments WHERE id = ?", req.CommentID).Scan(&likes, &dislikes)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Database error (display)"})
		return
	}

	// Return the updated counts.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]int{
		"likes":    likes,
		"dislikes": dislikes,
	})
}

func (ph *PostHandler) handleEditComment(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	if userID == "" {
		http.Error(w, "Unauthorized - Please log in to edit comment", http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	commentID, err := strconv.Atoi(r.FormValue("comment_id"))
	if err != nil {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(r.FormValue("post_id"))
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	content := r.FormValue("content")
	if content == "" {
		http.Error(w, "Comment cannot be empty", http.StatusBadRequest)
		return
	}

	// Verify comment ownership
	var commentUserID string
	err = utils.GlobalDB.QueryRow("SELECT user_id FROM comments WHERE id = ?", commentID).Scan(&commentUserID)
	if err != nil {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}

	if commentUserID != userID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Update comment
	_, err = utils.GlobalDB.Exec("UPDATE comments SET content = ? WHERE id = ?", content, commentID)
	if err != nil {
		log.Printf("Error updating comment: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/?id=%d", postID), http.StatusSeeOther)
}

// Add handler for deleting comments
func (ph *PostHandler) handleDeleteComment(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	if userID == "" {
		http.Error(w, "Unauthorized - Please log in to delete comments", http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	commentID, err := strconv.Atoi(r.FormValue("comment_id"))
	if err != nil {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(r.FormValue("post_id"))
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Verify comment ownership
	var commentUserID string
	err = utils.GlobalDB.QueryRow("SELECT user_id FROM comments WHERE id = ?", commentID).Scan(&commentUserID)
	if err != nil {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}

	if commentUserID != userID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Delete comment
	_, err = utils.GlobalDB.Exec("DELETE FROM comments WHERE id = ?", commentID)
	if err != nil {
		log.Printf("Error deleting comment: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Update post comment count
	_, err = utils.GlobalDB.Exec("UPDATE posts SET comments = comments - 1 WHERE id = ?", postID)
	if err != nil {
		log.Printf("Error updating comment count: %v", err)
	}

	http.Redirect(w, r, fmt.Sprintf("/?id=%d", postID), http.StatusSeeOther)
}
