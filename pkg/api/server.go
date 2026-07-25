package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/kubeinsights/kubeinsights/pkg/storage"
	"github.com/kubeinsights/kubeinsights/pkg/topology"
)

type Server struct {
	store    *storage.MemoryStore
	topology *topology.Discoverer
	mux      *http.ServeMux
}

func NewServer(store *storage.MemoryStore, topology *topology.Discoverer) *Server {
	s := &Server{
		store:    store,
		topology: topology,
		mux:      http.NewServeMux(),
	}
	s.routes()
	return s
}

func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	s.mux.HandleFunc("GET /api/traces", s.listTraces)
	s.mux.HandleFunc("GET /api/traces/", s.getTrace)
	s.mux.HandleFunc("GET /api/topology", s.getTopology)
}

func (s *Server) listTraces(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 50
	}
	writeJSON(w, http.StatusOK, s.store.List(limit))
}

func (s *Server) getTrace(w http.ResponseWriter, r *http.Request) {
	idText := strings.TrimPrefix(r.URL.Path, "/api/traces/")
	id, err := strconv.ParseUint(idText, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid trace id"})
		return
	}
	result, ok := s.store.Get(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "trace not found"})
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) getTopology(w http.ResponseWriter, r *http.Request) {
	devices, err := s.topology.Discover()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, devices)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
