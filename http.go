package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/eshop/config"
	"github.com/eshop/pkg/httperrors"
	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
)

func startHTTPServer(config *config.Configs, db *mongo.Database, redisClient *redis.Client) *http.Server {
	h := newHTTPHandler(config, db, redisClient)

	return &http.Server{
		Addr:    config.HTTPPort,
		Handler: h,
	}
}

func newHTTPHandler(config *config.Configs, db *mongo.Database, redisClient *redis.Client) http.Handler {
	r := mux.NewRouter()
	r.NewRoute().
		Methods(http.MethodGet).
		Path("/eshop/ping").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Pong"))
		})
	r.NewRoute().
		Methods(http.MethodPost).
		Path("/eshop/login").
		Handler(&authHTTPHandler{
			db:                     db,
			config:                 config,
			loginCredentialManager: &loginCredentialManagerImpl{},
		})
	r.NewRoute().
		Methods(http.MethodGet).
		Path("/eshop/products").
		Handler(&productHTTPHandler{
			db:             db,
			config:         config,
			productManager: &productManagerImpl{},
		})
	r.NewRoute().
		Methods(http.MethodPost).
		Path("/eshop/cart").
		Handler(&createCartHandler{
			db:             db,
			config:         config,
			redisClient:    redisClient,
			productManager: &productManagerImpl{},
			cartManager:    &cartManagerImpl{},
		})
	r.NewRoute().
		Methods(http.MethodGet).
		Path("/eshop/cart").
		Handler(&getCartHandler{
			db:          db,
			config:      config,
			redisClient: redisClient,
		})
	r.NewRoute().
		Methods(http.MethodPut).
		Path("/eshop/cart").
		Handler(&modifyCartHandler{
			db:             db,
			config:         config,
			redisClient:    redisClient,
			productManager: &productManagerImpl{},
			cartManager:    &cartManagerImpl{},
		})
	r.NewRoute().
		Methods(http.MethodPost).
		Path("/eshop/cart/checkout").
		Handler(&checkoutCartHandler{
			db:          db,
			config:      config,
			redisClient: redisClient,
		})
	r.NewRoute().
		Methods(http.MethodDelete).
		Path("/eshop/cart").
		Handler(&deleteCartHandler{
			db:          db,
			config:      config,
			redisClient: redisClient,
		})
	return r
}

type httpResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func handleHTTP(w http.ResponseWriter, req *http.Request, name string, f func(req *http.Request) ([]byte, error)) {
	response, err := f(req)
	if err != nil {
		err = errors.WithMessage(err, name)
		err = errors.WithMessage(err, "handle http")
		code, text := httperrors.GetCodeText(err)
		http.Error(w, text, code)
		log.Printf("Error : %v", err)
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(response)
	if err != nil {
		log.Printf("error while writing http respose :%v", err)
	}
}

func handleAuthHTTP(w http.ResponseWriter, req *http.Request, name string, f func(req *http.Request) ([]byte, error)) {
	response, err := f(req)
	if err != nil {
		handleHTTPError(w, name, err)
		return
	}

	var resp map[string]interface{}
	err = json.Unmarshal(response, &resp)
	if err != nil {
		handleHTTPError(w, name, err)
		return
	}
	if _, ok := resp["key"]; ok {
		if resp["key"].(string) == "Token" {
			http.SetCookie(w, &http.Cookie{
				Name:    resp["key"].(string),
				Value:   resp["tokenstring"].(string),
				Expires: time.Unix(int64(resp["expiresat"].(float64)), 0),
			})
		}
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(response)
	if err != nil {
		log.Printf("error while writing http respose :%v", err)
	}
}

func handleHTTPError(w http.ResponseWriter, name string, err error) {
	err = errors.WithMessage(err, name)
	err = errors.WithMessage(err, "handle Auth HTTP")
	code, text := httperrors.GetCodeText(err)
	http.Error(w, text, code)
	log.Printf("Error : %v", err)
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write([]byte(err.Error()))
	if err != nil {
		log.Printf("error while writing http respose :%v", err)
	}
}
