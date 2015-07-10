package server

import (
	"fmt"
	"net/http"
)

func (s *UAA) Error(w http.ResponseWriter, status int, message, errorType string) {
	output := fmt.Sprintf(`{"message":"%s","error":"%s"}`, message, errorType)

	w.WriteHeader(status)
	w.Write([]byte(output))
}

type validationError string

func (e validationError) Error() string {
	return string(e)
}
