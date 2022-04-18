package server

import "net/http"

func (s *Server) handlerHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handlerOpenAPI(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./openapi.json")
}
