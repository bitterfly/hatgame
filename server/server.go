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
	"github.com/bitterfly/go-chaos/hatgame/utils"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Server struct {
	Mux      *mux.Router
	Server   *http.Server
	DB       *gorm.DB
	Token    Token
	Games    map[uint]*containers.Game
	Mutex    *sync.RWMutex
	Upgrader websocket.Upgrader
}

func New(db *gorm.DB) *Server {
	return &Server{
		DB:    db,
		Mux:   mux.NewRouter(),
		Token: NewToken(32),
		Games: make(map[uint]*containers.Game),
		Mutex: &sync.RWMutex{},
		Upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			//TODO: Also fix this origin.
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
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
	s.Mux.HandleFunc("/host/{sessionToken}/{players}/{numWords}/{timer}", s.handleHost)
	s.Mux.HandleFunc("/join/{sessionToken}/{id}", s.handleJoin)
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
	log.Printf("Main, lol :D\n")
}

func (s *Server) handleWs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameIds, ok := vars["id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad id."))
		return
	}
	gameId, err := strconv.ParseUint(gameIds, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Id is not uint."))
		return
	}
	token, ok := vars["sessionToken"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad sessionToken."))
		return
	}

	payload, err := s.Token.VerifyToken(token)
	if err != nil {
		return
	}
	fmt.Printf("%d, %d\n", gameId, payload.Id)

	ws, err := s.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	s.Mutex.RLock()
	if _, ok := s.Games[uint(gameId)]; !ok {
		s.Mutex.RUnlock()
		return
	}

	s.Mutex.RUnlock()

	err = ws.WriteMessage(websocket.TextMessage, []byte("Hi!"))
	if err != nil {
		return
	}
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

	resp := map[string]interface{}{
		"sessionToken": token,
		"user":         dbUser,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
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
	payload, err := s.Token.CheckTokenRequest(w, r)
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
	err = database.UpdateUserPassword(s.DB, payload.Id, newPassowrd)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleHost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	players, err := utils.ParseInt(vars, "players")
	if err != nil {
		return
	}
	numWords, err := utils.ParseInt(vars, "numWords")
	if err != nil {
		return
	}
	timer, err := utils.ParseInt(vars, "timer")
	if err != nil {
		return
	}
	payload, err := s.Token.CheckTokenVars(vars)
	if err != nil {
		return
	}

	fmt.Printf("Players: %d, Words: %d, Timer: %d, HostId: %d\n", players, numWords, timer, payload.Id)

	user, err := database.GetUserByID(s.DB, payload.Id)
	if err != nil {
		fmt.Printf("Could not get user info for user: %d\n", payload.Id)
		return
	}

	s.Mutex.Lock()
	gameId := s.getGameId()
	s.Mutex.Unlock()

	game := containers.NewGame(gameId, *user, players, numWords, timer)
	ws, err := s.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	m, err := game.PutWs(payload.Id, ws)

	fmt.Printf("Sending: %s\n", m)
	s.Mutex.Lock()
	s.Games[gameId] = game
	s.Mutex.Unlock()

	err = game.NotifyAll(m)
	if err != nil {
		return
	}

	wordGuessed := make(chan struct{})
	msg := &containers.Message{}
	for {
		err := ws.ReadJSON(&msg)
		if err != nil {
			break
		}
		go msg.HandleMessage(ws, game, payload.Id, wordGuessed)
	}
}

func (s *Server) handleJoin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	gameId, err := utils.ParseUint(vars, "id")
	if err != nil {
		return
	}

	payload, err := s.Token.CheckTokenVars(vars)
	if err != nil {
		return
	}
	fmt.Printf("gameId: %d, userId: %d", gameId, payload.Id)

	user, err := database.GetUserByID(s.DB, payload.Id)
	if err != nil {
		return
	}

	s.Mutex.RLock()
	game, ok := s.Games[uint(gameId)]
	s.Mutex.RUnlock()
	if !ok {
		return
	}

	ws, err := s.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	m, err := game.PutAll(game.NumPlayers, *user, ws)
	if err != nil {
		return
	}

	fmt.Printf("Sending: %s\n", m)
	err = game.NotifyAll(m)
	if err != nil {
		return
	}

	wordGuessed := make(chan struct{})
	msg := &containers.Message{}
	for {
		err := ws.ReadJSON(&msg)
		if err != nil {
			fmt.Printf("Error reading json")
			break
		}
		go msg.HandleMessage(ws, game, payload.Id, wordGuessed)
	}
}
