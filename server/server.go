package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/bitterfly/go-chaos/hatgame/database"
	"github.com/bitterfly/go-chaos/hatgame/schema"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type Server struct {
	Mux     *mux.Router
	Server  *http.Server
	DB      *gorm.DB
	Token   Token
	Cookies map[string]struct{}
}

func New(db *gorm.DB) *Server {
	return &Server{DB: db, Mux: mux.NewRouter(), Token: NewToken(32), Cookies: make(map[string]struct{})}
}

func (s *Server) Connect(address string) error {
	s.Mux.HandleFunc("/", s.handleMain)
	s.Mux.HandleFunc("/login", s.handleUserLogin).Methods("POST")
	s.Mux.HandleFunc("/register", s.handleUserRegister).Methods("POST")
	s.Mux.HandleFunc("/user/id/{id}", s.handleUserShow).Methods("GET")
	s.Mux.HandleFunc("/user/password", s.handleUserPassword).Methods("POST")
	s.Mux.Use(mux.CORSMethodMiddleware(s.Mux))
	log.Printf("Starting server on %s\n", address)

	//TODO: fix the allowed origins
	allowedOrigins := handlers.AllowedOrigins([]string{"*"})
	allowedMethods := handlers.AllowedMethods([]string{"POST", "OPTIONS", "GET"})
	allowedHeaders := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})

	if err := http.ListenAndServe(
		address,
		handlers.LoggingHandler(os.Stderr, handlers.CORS(
			allowedOrigins,
			allowedMethods,
			allowedHeaders)(s.Mux))); err != nil {
		return fmt.Errorf("error connecting to server %s: %w", address, err)
	}
	return nil
}

func (s *Server) handleMain(w http.ResponseWriter, r *http.Request) {
	log.Printf("Main :D\n")
}

func (s *Server) handleUserLogin(w http.ResponseWriter, r *http.Request) {
	user, err := schema.ParseUser(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad user json."))
		return
	}
	dbUser, err := database.GetUserByEmail(s.DB, user.Email)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Wrong email or password."))
		return
	}
	if dbUser.Password != user.Password {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Wrong email or password."))
		return
	}

	token, err := s.Token.CreateToken(dbUser.ID, 15)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte("Could not create authentication token."))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(token))
	s.Cookies[token] = struct{}{}
}

func (s *Server) handleUserRegister(w http.ResponseWriter, r *http.Request) {
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
	user, err := database.GetUserByID(s.DB, uint(id_u))
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("No user with id: %d.", id_u)))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func (s *Server) handleUserPassword(w http.ResponseWriter, r *http.Request) {
	token := ExtractToken(r)
	fmt.Printf("%s\n", token)
	payload, err := s.Token.VerifyToken(token)
	fmt.Printf("%v, %v\n", payload, err)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(fmt.Sprintf("%s", err)))
		return
	}
	user, err := schema.ParseUser(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad user json."))
		return
	}
	err = database.UpdateUserPassword(s.DB, payload.Id, user.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
