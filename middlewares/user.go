package middlewares

import (
	"context"
	"net/http"

	"github.com/Dimitriy14/image-resizing/logger"
	"github.com/Dimitriy14/image-resizing/services/common"
	"github.com/google/uuid"
)

// headerUID is user ID header
const headerUID = "UID"

func CheckUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, isUser := getRequestUID(r)
		if !isUser {
			logger.Log.Errorf("user is missing in request")
			common.SendError(w, http.StatusUnauthorized, "user was not specified", nil)
			return
		}

		id, err := uuid.Parse(userID)
		if err != nil {
			logger.Log.Errorf("parsing user id: %s", err)
			common.SendError(w, http.StatusUnauthorized, "invalid user id", nil)
			return
		}

		ctx := context.WithValue(r.Context(), common.UserID, id)
		next.ServeHTTP(w, r.WithContext(ctx))
		return

	})
}

// getRequestUID returns UID header from request
func getRequestUID(r *http.Request) (string, bool) {
	val := r.Header.Get(headerUID)
	return val, val != ""
}
