package postcontollers

import (
	"encoding/json"
	"net/http"

	"forum/utils"
)

func (ph *PostHandler) handleEditPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from session
	userID := r.Context().Value("userID").(string)

	var req struct {
		PostID  int    `json:"post_id"`
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Verify post ownership
	var postUserID string
	err := utils.GlobalDB.QueryRow("SELECT user_id FROM posts WHERE id = ?", req.PostID).Scan(&postUserID)
	if err != nil || postUserID != userID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Update post
	_, err = utils.GlobalDB.Exec("UPDATE posts SET content = ? WHERE id = ?", req.Content, req.PostID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
