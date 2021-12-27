package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/bitterfly/go-chaos/hatgame/schema"
	"gorm.io/gorm"
)

type Server struct {
	Mux    *http.ServeMux
	Server *http.Server
	DB     *gorm.DB
}

func New(db *gorm.DB, address, port string) *Server {
	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", address, port),
		Handler: mux,
	}
	return &Server{DB: db, Mux: mux, Server: server}
}

func (s *Server) Connect() error {
	s.Mux.HandleFunc("/", s.handleMain)
	s.Mux.HandleFunc("/new", s.handleNew)
	log.Printf("Starting server on %s\n", s.Server.Addr)

	if err := s.Server.ListenAndServe(); err != nil {
		return fmt.Errorf("error connecting to server %s: %w", s.Server.Addr, err)
	}
	return nil
}

func (s *Server) handleMain(w http.ResponseWriter, r *http.Request) {
	log.Printf("Main :D\n")
}

func (s *Server) handleNew(w http.ResponseWriter, r *http.Request) {
	user, err := ParseUser(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad user. Nutty boi."))
		return
	}
	id, err := AddUser(s.DB, user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("%d", id)))
}

func ParseUser(data io.ReadCloser) (*schema.Users, error) {
	var user schema.Users
	if err := json.NewDecoder(data).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

func AddUser(db *gorm.DB, user *schema.Users) (uint, error) {
	if err := db.Create(user).Error; err != nil {
		return 0, err
	}
	return user.ID, nil
}
