package server

import (
	"fmt"
	"net/http"
)

func (s *UAA) notFound(w http.ResponseWriter, message string) {
	output := fmt.Sprintf(`{"message":"%s","error":"scim_resource_not_found"}`, message)

	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(output))
}
