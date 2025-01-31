package utils

import "database/sql"

func ToggleReaction(db *sql.DB, userID string, postID int, reactionType string) (map[string]int, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Check existing reaction
	var exists bool
	err = tx.QueryRow(`
        SELECT EXISTS(
            SELECT 1 FROM reaction 
            WHERE user_id = ? AND post_id = ? AND like = ?
        )
    `, userID, postID, reactionType == "like").Scan(&exists)
	if err != nil {
		return nil, err
	}

	if exists {
		// Remove reaction
		_, err = tx.Exec(`
            DELETE FROM reaction 
            WHERE user_id = ? AND post_id = ?
        `, userID, postID)
	} else {
		// Add new reaction
		_, err = tx.Exec(`
            INSERT INTO reaction (user_id, post_id, like)
            VALUES (?, ?, ?)
        `, userID, postID, reactionType == "like")
	}
	if err != nil {
		return nil, err
	}

	// Get updated counts
	var likes, dislikes int
	err = tx.QueryRow(`
        SELECT 
            SUM(CASE WHEN like = 1 THEN 1 ELSE 0 END),
            SUM(CASE WHEN like = 0 THEN 1 ELSE 0 END)
        FROM reaction
        WHERE post_id = ?
    `, postID).Scan(&likes, &dislikes)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return map[string]int{
		"likes":    likes,
		"dislikes": dislikes,
	}, nil
}
