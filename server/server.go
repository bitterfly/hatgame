package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/bitterfly/go-chaos/hatgame/database"
	"github.com/bitterfly/go-chaos/hatgame/schema"
	"github.com/bitterfly/go-chaos/hatgame/server/containers"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Server struct {
	Mux    *mux.Router
	Server *http.Server
	DB     *gorm.DB
	Token  Token
	Games  map[uint]containers.Game
	Mutex  *sync.RWMutex
}

func New(db *gorm.DB) *Server {
	return &Server{
		DB:    db,
		Mux:   mux.NewRouter(),
		Token: NewToken(32),
		Games: make(map[uint]containers.Game),
		Mutex: &sync.RWMutex{},
	}
}

func (s *Server) getGameId() uint {
	var m uint
	for k := range s.Games {
		if k > m {
			m = k
		}
	}
	return m + 1
}

func (s *Server) Connect(address string) error {
	s.Mux.HandleFunc("/", s.handleMain)
	s.Mux.HandleFunc("/login", s.handleUserLogin).Methods("POST")
	s.Mux.HandleFunc("/register", s.handleUserRegister).Methods("POST")
	s.Mux.HandleFunc("/user/id/{id}", s.handleUserShow).Methods("GET")
	s.Mux.HandleFunc("/user/password", s.handleUserPassword).Methods("POST")
	s.Mux.HandleFunc("/host", s.handleHost).Methods("POST")
	s.Mux.HandleFunc("/game/{id}", s.handleJoin).Methods("POST")
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
	user, err := containers.ParseUser(r.Body)
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
	f, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	fmt.Printf("o: %s\nh: %v\ndb: %v\n", user.Password, f, dbUser.Password)

	if err := bcrypt.CompareHashAndPassword(dbUser.Password, []byte(user.Password)); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		log.Printf("%s\n", err.Error())
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
	idU, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Id is not uint."))
		return
	}
	user, err := database.GetUserByID(s.DB, uint(idU))
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("No user with id: %d.", idU)))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func (s *Server) handleUserPassword(w http.ResponseWriter, r *http.Request) {
	payload, err := s.Token.CheckToken(w, r)
	if err != nil {
		return
	}

	user, err := containers.ParseUser(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad user json."))
		return
	}

	newPassowrd, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Could not encrypt password."))
		return
	}
	fmt.Printf("Password: %s Writing to database %d, %v\n", user.Password, payload.Id, newPassowrd)
	err = database.UpdateUserPassword(s.DB, payload.Id, newPassowrd)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleHost(w http.ResponseWriter, r *http.Request) {
	payload, err := s.Token.CheckToken(w, r)
	if err != nil {
		return
	}

	host, err := containers.ParseHost(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad user json."))
		return
	}

	s.Mutex.Lock()
	gameId := s.getGameId()
	game := containers.NewGame(gameId, payload.Id, host.Players, host.Timer)
	s.Games[gameId] = game
	s.Mutex.Unlock()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(game)
}

func (s *Server) handleJoin(w http.ResponseWriter, r *http.Request) {
	payload, err := s.Token.CheckToken(w, r)
	if err != nil {
		return
	}

	vars := mux.Vars(r)
	gameId, ok := vars["id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad game id."))
		return
	}

	gameIdU, err := strconv.ParseUint(gameId, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Game id is not uint."))
		return
	}

	s.Mutex.RLock()
	game, ok := s.Games[uint(gameIdU)]
	s.Mutex.RUnlock()
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Game with id: %d does not exists.", gameIdU)))
		return
	}

	err = game.Put(game.NumPlayers, payload.Id)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(fmt.Sprintf("Cannot join game with id %d: %s", gameIdU, err.Error())))
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(game)
}
