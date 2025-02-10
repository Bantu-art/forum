package postcontollers

import (
	"net/http"

	"forum/utils"
)

func (ph *PostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/create":
		switch r.Method {
		case http.MethodGet:
			ph.authMiddleware(ph.displayCreateForm).ServeHTTP(w, r)
		case http.MethodPost:
			ph.authMiddleware(ph.handleCreatePost).ServeHTTP(w, r)
		default:
			utils.RenderErrorPage(w, http.StatusMethodNotAllowed, utils.ErrMethodNotAllowed)
		}
	case "/react":
		if r.Method == http.MethodPost {
			ph.authMiddleware(ph.handleReactions).ServeHTTP(w, r)
		} else {
			utils.RenderErrorPage(w, http.StatusMethodNotAllowed, utils.ErrMethodNotAllowed)
		}
	case "/":
		if r.Method == http.MethodGet {
			if r.URL.Query().Get("id") != "" {
				ph.handleSinglePost(w, r)
			} else {
				ph.handleGetPosts(w, r)
			}
		}
	case "/comment":
		if r.Method == http.MethodPost {
			ph.authMiddleware(ph.handleComment).ServeHTTP(w, r)
		} else {
			utils.RenderErrorPage(w, http.StatusMethodNotAllowed, utils.ErrMethodNotAllowed)
		}
	case "/commentreact":
		if r.Method == http.MethodPost {
			ph.authMiddleware(ph.handleCommentReactions).ServeHTTP(w, r)
		} else {
			utils.RenderErrorPage(w, http.StatusMethodNotAllowed, utils.ErrMethodNotAllowed)
		}
	case "/edit-post":
		if r.Method == http.MethodPost {
			ph.authMiddleware(ph.handleEditPost).ServeHTTP(w, r)
		} else {
			utils.RenderErrorPage(w, http.StatusMethodNotAllowed, utils.ErrMethodNotAllowed)
		}
	case "/delete-post":
		if r.Method == http.MethodPost {
			ph.authMiddleware(ph.handleDeletePost).ServeHTTP(w, r)
		} else {
			utils.RenderErrorPage(w, http.StatusMethodNotAllowed, utils.ErrMethodNotAllowed)
		}
		
	default:
		utils.RenderErrorPage(w, http.StatusNotFound, utils.ErrPageNotFound)
	}
}
