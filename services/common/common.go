package common

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/google/uuid"

	"github.com/Dimitriy14/image-resizing/logger"
)

type UserKey string

const (
	contentType     = "Content-Type"
	applicationJSON = "application/json; charset=utf-8"

	// UserID is used for checking user id in middleware
	UserID UserKey = "UserKey"
)

var createError = func(msg string) interface{} {
	return ErrorMessage{Error{msg}}
}

//Error contains the message about error
type Error struct {
	Message string `json:"message"`
}

//ErrorMessage contains the error
type ErrorMessage struct {
	Error Error `json:"error"`
}

// RenderJSONCreated is used for rendering JSON response body when new resource has been created
func RenderJSONCreated(w http.ResponseWriter, response interface{}) {
	data, err := json.Marshal(response)
	if err != nil {
		SendInternalServerError(w, "failed to marshal response", err)
		return
	}
	render(w, http.StatusCreated, data)
}

func render(w http.ResponseWriter, code int, response []byte) {
	w.Header().Set(contentType, applicationJSON)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(code)
	_, err := w.Write(response)
	if err != nil {
		logger.Log.Debugf("Failed to Write response %v (err: %v)", response, err)
	}
}

// ReadRequestJSONBodyToStruct reads request body to a particular structure
func ReadRequestJSONBodyToStruct(r *http.Request, v interface{}) (err error) {
	err = json.NewDecoder(r.Body).Decode(v)
	CloseWithErrCheck(r.Body, "request")
	return
}

// SendConflictError sends Conflict Status
func SendConflictError(w http.ResponseWriter, message string) {
	SendError(w, http.StatusConflict, message, nil)
}

// SendNotFound sends 404 response
func SendNotFound(w http.ResponseWriter, message string, args ...interface{}) {
	var err error
	if len(args) > 0 {
		err = fmt.Errorf(message, args...)
	} else {
		err = errors.New(message)
	}

	SendError(w, http.StatusNotFound, err.Error(), err)
}

// CloseWithErrCheck runs f.Close() checking/logging returned error.
func CloseWithErrCheck(f io.Closer, name string) {
	err := f.Close()
	if err != nil {
		if logger.Log != nil {
			logger.Log.Errorf("Failed to close %s (err: %v)", name, err)
		} else {
			log.Printf("ERROR: Failed to close %s (err: %s)", name, err)
		}
	}
}

// SendError writes a defined string as an error message
// with appropriate headers to the HTTP response
func SendError(w http.ResponseWriter, code int, message string, err error) {
	if err != nil {
		logger.Log.Errorf(message, "%v", err)
	}
	if message == "" {
		message = http.StatusText(code)
	}
	data, err := json.Marshal(createError(message))
	if err != nil {
		logger.Log.Errorf("", message, "helpers.SendError: %v", err)
	}
	render(w, code, data)
}

// SendInternalServerError sends Internal Server Error Status and logs an error if it exists
func SendInternalServerError(w http.ResponseWriter, message string, err error) {
	SendError(w, http.StatusInternalServerError, message, err)
}

// GetUserIDFromCtx retrieves user id from context
func GetUserIDFromCtx(ctx context.Context) uuid.UUID {
	id, ok := ctx.Value(UserID).(uuid.UUID)
	if !ok {
		logger.Log.Warnf("User id is missing in ctx")
	}
	return id
}
