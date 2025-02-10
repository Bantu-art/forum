package postcontollers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"forum/utils"
)

func (ph *PostHandler) handleReactions(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	if userID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	var req struct {
		PostID int `json:"post_id"`
		Like   int `json:"like"` // 1 for like, 0 for dislike
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

	// Check if the user already has a reaction
	var existingLike int
	err := utils.GlobalDB.QueryRow("SELECT like FROM reaction WHERE user_id = ? AND post_id = ?", userID, req.PostID).Scan(&existingLike)
	if err != nil && err != sql.ErrNoRows {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Database error"})
		return
	}

	if err == sql.ErrNoRows {
		// Insert new reaction
		_, err = utils.GlobalDB.Exec("INSERT INTO reaction (user_id, post_id, like) VALUES (?, ?, ?)", userID, req.PostID, req.Like)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Database error"})
			return
		}
	} else {
		if existingLike == req.Like {
			// User is unliking or undisliking
			_, err = utils.GlobalDB.Exec("DELETE FROM reaction WHERE user_id = ? AND post_id = ?", userID, req.PostID)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": "Database error"})
				return
			}
		} else {
			// Update existing reaction
			_, err = utils.GlobalDB.Exec("UPDATE reaction SET like = ? WHERE user_id = ? AND post_id = ?", req.Like, userID, req.PostID)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": "Database error"})
				return
			}
		}
	}

	// Fetch updated like and dislike counts
	var likes, dislikes int
	err = utils.GlobalDB.QueryRow("SELECT likes, dislikes FROM posts WHERE id = ?", req.PostID).Scan(&likes, &dislikes)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Database error"})
		return
	}

	// Return updated counts as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]int{
		"likes":    likes,
		"dislikes": dislikes,
	})
}
