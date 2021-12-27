package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/bitterfly/go-chaos/hatgame/database"
	"github.com/bitterfly/go-chaos/hatgame/schema"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type Server struct {
	Mux    *mux.Router
	Server *http.Server
	DB     *gorm.DB
}

func New(db *gorm.DB, address, port string) *Server {
	m := mux.NewRouter()
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", address, port),
		Handler: m,
	}
	return &Server{DB: db, Mux: m, Server: server}
}

func (s *Server) Connect() error {
	s.Mux.HandleFunc("/", s.handleMain)
	s.Mux.HandleFunc("/register", s.handleUserRegister)
	s.Mux.HandleFunc("/user/{id}", s.handleUserShow)
	log.Printf("Starting server on %s\n", s.Server.Addr)

	if err := s.Server.ListenAndServe(); err != nil {
		return fmt.Errorf("error connecting to server %s: %w", s.Server.Addr, err)
	}
	return nil
}

func (s *Server) handleMain(w http.ResponseWriter, r *http.Request) {
	log.Printf("Main :D\n")
}

func (s *Server) handleUserRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Not post request."))
		return
	}

	user, err := schema.ParseUser(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad user json."))
		return
	}
	id, err := database.AddUser(s.DB, user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("%d", id)))
}

func (s *Server) handleUserShow(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Not get request."))
		return
	}
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad id."))
		return
	}
	id_u, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Id is not uint."))
		return
	}
	user, err := database.GetUser(s.DB, uint(id_u))
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("No user with id: %d.", id_u)))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}
