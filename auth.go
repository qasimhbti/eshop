package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/eshop/config"
	"github.com/eshop/pkg/httperrors"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
)

type authHTTPHandler struct {
	db                     *mongo.Database
	config                 *config.Configs
	loginCredentialManager interface {
		get(userName string, db *mongo.Database) (*loginCred, error)
	}
}

func (h *authHTTPHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	handleAuthHTTP(w, req, "EShop-Login", h.handle)
}

type loginResponse struct {
	Key         string `json:"key"`
	TokenString string `json:"tokenstring"`
	Status      string `json:"status"`
	ExpiresAt   int64  `json:"expiresat"`
}

type claims struct {
	UserName string `json:"username"`
	jwt.StandardClaims
}

func (h *authHTTPHandler) handle(r *http.Request) ([]byte, error) {
	err := r.ParseForm()
	if err != nil {
		return nil, errors.WithMessage(
			httperrors.WithCode(err, http.StatusBadRequest),
			"get login data",
		)
	}
	userName := r.PostForm.Get("username")
	password := r.PostForm.Get("password")

	if userName == "" || password == "" {
		err := errors.New("username or password is empty")
		return nil, errors.WithMessage(
			httperrors.WithCode(err, http.StatusBadRequest),
			"get login data",
		)
	}

	loginCreds, err := h.loginCredentialManager.get(userName, h.db)
	if err != nil {
		return nil, errors.WithMessage(
			httperrors.WithCode(err, http.StatusUnprocessableEntity),
			"get credential data",
		)
	}

	if password != loginCreds.Password {
		err := errors.New("password is invalid")
		return nil, errors.WithMessage(
			httperrors.WithCode(err, http.StatusUnauthorized),
			"invalid credentials",
		)
	}

	//expiration time of token
	expirationTime := time.Now().Add(15 * time.Minute)
	clms := &claims{
		UserName: userName,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, clms)

	tokenString, err := token.SignedString([]byte(h.config.JWTAccessSecretKey))
	if err != nil {
		return nil, errors.WithMessage(
			httperrors.WithCode(err, http.StatusUnprocessableEntity),
			"generating signed token",
		)
	}

	log.Printf("user: %s successfully login\n", loginCreds.UserName)
	resp, err := json.Marshal(&loginResponse{
		Key:         "Token",
		TokenString: tokenString,
		Status:      "Successfully Generated",
		ExpiresAt:   expirationTime.Unix(),
	})
	if err != nil {
		return nil, errors.WithMessage(
			httperrors.WithCode(err, http.StatusUnprocessableEntity),
			"JSON Marshal",
		)
	}
	return resp, nil
}

// ProtectedEndpoint --
func protectedEndpoint(jwtKey string, req *http.Request) (*claims, error) {
	cookie, err := req.Cookie("Token")
	if err != nil {
		if err == http.ErrNoCookie {
			// If the cookie is not set, return an unauthorized status
			log.Println("Cookie not set")
		}
		// For any other type of error, return a bad request status
		return nil, errors.WithMessage(err, "Invalid Request")
	}

	clms := &claims{}
	tkn, err := jwt.ParseWithClaims(cookie.Value, clms, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtKey), nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return nil, errors.WithMessage(err, "Invalid Signature")
		}
		return nil, errors.WithMessage(err, "Access Denied-Please check the access token")
	}

	if !tkn.Valid {
		return nil, errors.WithMessage(err, "Invalid Token")
	}
	return clms, nil
}
