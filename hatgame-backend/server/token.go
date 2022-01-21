package server

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var (
	ErrInvalidToken = errors.New("token is invalid")
	ErrExpiredToken = errors.New("token has expired")
)

type Payload struct {
	Id      uint
	Expires int64
}

func (payload *Payload) Valid() error {
	if time.Now().After(time.Unix(payload.Expires, 0)) {
		return ErrExpiredToken
	}
	return nil
}

type Token struct {
	secretKey string
}

func NewToken(n int) Token {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return Token{secretKey: string(s)}
}

func (t *Token) CreateToken(id uint, duration time.Duration) (string, error) {
	payload := Payload{Id: id, Expires: time.Now().Add(time.Minute * 15).Unix()}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, &payload)
	signedToken, err := jwtToken.SignedString([]byte(t.secretKey))
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

func (t *Token) CheckTokenRequest(w http.ResponseWriter, r *http.Request) (*Payload, error) {
	token := ExtractToken(r)
	payload, err := t.VerifyToken(token)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(fmt.Sprintf("%s", err)))
		return nil, err
	}
	return payload, nil
}

func ExtractToken(r *http.Request) string {
	bearToken := r.Header.Get("Authorization")
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}

func (t *Token) VerifyToken(token string) (*Payload, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, ErrInvalidToken
		}
		return []byte(t.secretKey), nil
	}

	jwtToken, err := jwt.ParseWithClaims(token, &Payload{}, keyFunc)
	if err != nil {
		verr, ok := err.(*jwt.ValidationError)
		if ok && errors.Is(verr.Inner, ErrExpiredToken) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	payload, ok := jwtToken.Claims.(*Payload)
	if !ok {
		return nil, ErrInvalidToken
	}

	return payload, nil
}

func (t *Token) CheckTokenVars(vars map[string]string) (*Payload, error) {
	token, ok := vars["sessionToken"]
	if !ok {
		return nil, fmt.Errorf("missing parameter for sessionTokne")
	}

	payload, err := t.VerifyToken(token)
	if err != nil {
		return nil, err
	}
	return payload, nil
}
