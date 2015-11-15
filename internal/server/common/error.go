package common

import (
	"fmt"
	"net/http"
)

func Error(w http.ResponseWriter, status int, message, errorType string) {
	output := fmt.Sprintf(`{"error_description":%q,"error":%q}`, message, errorType)

	w.WriteHeader(status)
	w.Write([]byte(output))
}
