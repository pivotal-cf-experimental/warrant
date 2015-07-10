package groups

import (
	"github.com/gorilla/mux"
	"github.com/pivotal-cf-experimental/warrant/internal/server/domain"
)

func NewRouter(groups *domain.Groups, tokens *domain.Tokens) *mux.Router {
	router := mux.NewRouter()

	router.Handle("/Groups", createHandler{groups, tokens}).Methods("POST")
	router.Handle("/Groups", listHandler{groups}).Methods("GET")
	router.Handle("/Groups/{guid}", getHandler{groups, tokens}).Methods("GET")
	router.Handle("/Groups/{guid}", deleteHandler{groups, tokens}).Methods("DELETE")

	return router
}
