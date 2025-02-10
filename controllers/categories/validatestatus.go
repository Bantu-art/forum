package categories


import (
	"net/http"

	"forum/utils"
)

func (ch *CategoryHandler) checkAuthStatus(r *http.Request) bool {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return false
	}
	_, err = utils.ValidateSession(utils.GlobalDB, cookie.Value)
	return err == nil
}
