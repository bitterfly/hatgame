package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
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

type Game struct {
	Players map[uint]*websocket.Conn
	State   *containers.Game
}

type Server struct {
	Mux      *mux.Router
	Server   *http.Server
	DB       *gorm.DB
	Token    Token
	Games    map[uint]Game
	Mutex    *sync.RWMutex
	Upgrader websocket.Upgrader
}

func New(db *gorm.DB) *Server {
	return &Server{
		DB:    db,
		Mux:   mux.NewRouter(),
		Token: NewToken(32),
		Games: make(map[uint]Game),
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
	authRouter := s.Mux.NewRoute().Subrouter()
	authRouter.Use(s.authHandler)
	authRouter.HandleFunc("/api/user/id/{id}", s.handleUserShow).Methods("GET")
	authRouter.HandleFunc("/api/game/id/{id}", s.handleGameShow).Methods("POST")
	authRouter.HandleFunc("/api/user/change", s.handleUserChange).Methods("POST")
	authRouter.HandleFunc("/api/user", s.handleUserGet).Methods("POST")
	authRouter.HandleFunc("/api/stat", s.handleStat).Methods("GET")
	authRouter.HandleFunc("/api/recommend", s.handleRecommend).Methods("POST")

	s.Mux.HandleFunc("/api/", s.handleMain)
	s.Mux.HandleFunc("/api/login", s.handleUserLogin).Methods("POST")
	s.Mux.HandleFunc("/api/register", s.handleUserRegister).Methods("POST")
	s.Mux.HandleFunc("/api/host/{sessionToken}/{players}/{numWords}/{timer}", s.handleHost)
	s.Mux.HandleFunc("/api/join/{sessionToken}/{id}", s.handleJoin)
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

func (s *Server) authHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, err := s.Token.CheckTokenRequest(w, r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "id", payload.Id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) handleMain(w http.ResponseWriter, r *http.Request) {
	//log.Printf("Main, lol :D\n")
}

func (s *Server) handleUserLogin(w http.ResponseWriter, r *http.Request) {
	user, err := containers.ParseUser(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad user json."))
		return
	}
	dbUser, derr := database.GetUserByEmail(s.DB, user.Email)
	if derr != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Wrong email or password."))
		return
	}
	if err := bcrypt.CompareHashAndPassword(dbUser.Password, []byte(user.Password)); err != nil {
		log.Printf("%s\n", err.Error())
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Wrong email or password."))
		return
	}

	token, err := s.Token.CreateToken(dbUser.ID, 15)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
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
	user, err := containers.ParseUser(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad user json."))
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Could not encode password"))
		return
	}

	schemaUser := &schema.User{
		Email:    user.Email,
		Password: hashedPassword,
		Username: user.Username,
	}

	id, derr := database.AddUser(s.DB, schemaUser)
	if derr != nil {
		if derr.ErrorType == database.ConflictError {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(derr.Error()))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(derr.Error()))
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
	user, derr := database.GetUserByID(s.DB, uint(idU))
	if derr != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("No user with id: %d.", idU)))
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func (s *Server) handleStat(w http.ResponseWriter, r *http.Request) {
	id, ok := r.Context().Value("id").(uint)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	stat, derr := database.GetUserStatistics(s.DB, id)
	if derr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(stat)
}

func (s *Server) handleRecommend(w http.ResponseWriter, r *http.Request) {
	id, ok := r.Context().Value("id").(uint)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	nStr := r.URL.Query().Get("n")
	if nStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Missing required query param \"n\"."))
		return
	}
	n, err := strconv.Atoi(nStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Could not parse query param \"n\" as integer."))
		return
	}

	result, derr := database.RecommendWord(s.DB, n, id)
	if derr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

func (s *Server) handleUserGet(w http.ResponseWriter, r *http.Request) {
	id, ok := r.Context().Value("id").(uint)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	dbUser, derr := database.GetUserByID(s.DB, id)
	if derr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Could not fetch from database."))
		return
	}

	resp := map[string]interface{}{
		"sessionToken": ExtractToken(r),
		"user":         dbUser,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleGameShow(w http.ResponseWriter, r *http.Request) {
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

	s.Mutex.RLock()
	defer s.Mutex.RUnlock()
	if _, ok := s.Games[uint(idU)]; !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(s.Games[uint(idU)])
}

func (s *Server) handleUserChange(w http.ResponseWriter, r *http.Request) {
	id, ok := r.Context().Value("id").(uint)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	user, err := containers.ParseUser(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad user json."))
		return
	}

	strippedPassword := strings.TrimSpace(user.Password)
	if strippedPassword != "" {
		newPassowrd, err := bcrypt.GenerateFromPassword([]byte(strippedPassword), bcrypt.DefaultCost)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Could not encrypt password."))
			return
		}
		derr := database.UpdateUser(s.DB, id, newPassowrd, user.Username)
		if derr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		derr := database.UpdateUserUsername(s.DB, id, user.Username)
		if derr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleHost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	numPlayers, err := utils.ParseInt(vars, "players")
	if err != nil {
		log.Printf("[handleHost] Could not parse \"player\" var: %s", err.Error())
		return
	}
	numWords, err := utils.ParseInt(vars, "numWords")
	if err != nil {
		log.Printf("[handleHost] Could not parse \"numWords\" var: %s", err.Error())
		return
	}
	timer, err := utils.ParseInt(vars, "timer")
	if err != nil {
		log.Printf("[handleHost] Could not parse \"timer\" var: %s", err.Error())
		return
	}
	payload, err := s.Token.CheckTokenVars(vars)
	if err != nil {
		log.Printf("[handleHost] Could not validate token: %s", err.Error())
		return
	}

	user, derr := database.GetUserByID(s.DB, payload.Id)
	if derr != nil {
		log.Printf("[handleHost] Could not get user info for user: %d\n", payload.Id)
		return
	}

	s.Mutex.Lock()
	gameId := s.getGameId()
	s.Mutex.Unlock()

	game := containers.NewGame(gameId, *user, numPlayers, numWords, timer)
	ws, err := s.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[handleHost] Could not upgrade to ws")
		return
	}
	err = game.PutWs(payload.Id, ws)
	if err != nil {
		log.Printf("[handleHost] %s", err.Error())
		return
	}
	m, err := containers.CreateMessage("game", game)
	if err != nil {
		log.Printf("[handleHost] %s", err.Error())
		return
	}

	players := make(map[uint]*websocket.Conn)
	players[payload.Id] = ws

	s.Mutex.Lock()
	s.Games[gameId] = Game{Players: players, State: game}
	s.Mutex.Unlock()

	err = game.NotifyAll(m)
	if err != nil {
		log.Printf("[handleHost] Could not notify all players: %s", err.Error())
		return
	}
	s.listen(ws, game, payload.Id)
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	delete(s.Games, gameId)
	derr = database.AddGame(s.DB, game)
	if derr != nil {
		log.Printf("Error when inserting game to database: %s", err.Error())
	}
}

func (s *Server) handleJoin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	gameId, err := utils.ParseUint(vars, "id")
	if err != nil {
		log.Printf("[handleJoin] Could not parse \"gameId\" var: %s", err.Error())
		return
	}

	payload, err := s.Token.CheckTokenVars(vars)
	if err != nil {
		log.Printf("[handleJoin] Could not validate token: %s", err.Error())
		return
	}

	user, derr := database.GetUserByID(s.DB, payload.Id)
	if derr != nil {
		log.Printf("[handleJoin] Could not get user info for user: %d\n", payload.Id)
		return
	}

	s.Mutex.RLock()
	game, ok := s.Games[uint(gameId)]
	s.Mutex.RUnlock()
	if !ok {
		log.Printf("[handleJoin] No game with id: %d\n", gameId)
		return
	}

	ws, err := s.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[handleJoin] Could not upgrade to ws: %s", err.Error())
		return
	}

	if err := game.State.Put(game.State.NumPlayers, *user, ws); err != nil {
		log.Printf("[handleJoin] %s", err.Error())
		resp, err := containers.CreateMessage("error", err.Error())
		if err != nil {
			log.Printf("[handleJoin] %s", err.Error())
		}
		err = ws.WriteMessage(websocket.TextMessage, resp)
		if err != nil {
			log.Printf("[handleJoin] %s", err.Error())
		}
		return
	}

	game.Players[user.ID] = ws

	m, err := containers.CreateMessage("game", game)
	if err != nil {
		log.Printf("[handleJoin] %s", err.Error())
		return
	}

	if err := game.State.NotifyAll(m); err != nil {
		log.Printf("[handleJoin] %s", err.Error())
		return
	}

	s.listen(ws, game.State, payload.Id)
}

func (s *Server) listen(ws *websocket.Conn, game *containers.Game, id uint) {
	msg := &containers.Message{}
	errors := make(chan error)
	defer close(errors)
	go HandleErrors(errors)
	message := make(chan *containers.Message, 1)
	defer close(message)

	go func(ws *websocket.Conn) {
		for {
			err := ws.ReadJSON(&msg)
			if err != nil {
				return
			}
			message <- msg
		}
	}(ws)

	for {
		select {
		case _, ok := <-game.Process.GameEnd:
			if !ok {
				ws.Close()
				return
			}
		case msg := <-message:
			go msg.HandleMessage(ws, game, id, errors)
		}
	}
}

func HandleErrors(errors chan error) {
	for {
		err, ok := <-errors
		if !ok {
			return
		}
		log.Printf("ERROR: %s.\n", err)
	}
}
