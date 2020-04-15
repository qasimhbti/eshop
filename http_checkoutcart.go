package main

import (
	"encoding/json"
	"net/http"

	"github.com/eshop/config"
	"github.com/eshop/pkg/httperrors"
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
)

type checkoutCartHandler struct {
	db          *mongo.Database
	config      *config.Configs
	redisClient *redis.Client
}

func (h *checkoutCartHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	handleHTTP(w, req, "Checkout Cart Handler", h.handle)
}

func (h *checkoutCartHandler) handle(req *http.Request) ([]byte, error) {
	claims, err := protectedEndpoint(h.config.JWTAccessSecretKey, req)
	if err != nil {
		return nil, errors.WithMessage(
			httperrors.WithCode(err, http.StatusForbidden),
			"Invalid Login Credentials",
		)
	}

	userCart, err := h.redisClient.Get(claims.UserName).Result()
	if err != nil {
		if err != redis.Nil {
			return nil, errors.WithMessage(
				httperrors.WithCode(err, http.StatusUnprocessableEntity),
				"redis-getting user cart",
			)
		}
	}

	_, err = h.redisClient.Del(claims.UserName).Result()
	if err != nil {
		if err != redis.Nil {
			return nil, errors.WithMessage(
				httperrors.WithCode(err, http.StatusBadRequest),
				"redis-deleting old cart",
			)
		}
	}

	resp := &httpResponse{
		Status:  "CHECKOUT",
		Message: userCart,
	}
	response, err := json.Marshal(resp)
	if err != nil {
		return nil, errors.WithMessage(
			httperrors.WithCode(err, http.StatusUnprocessableEntity),
			"cart-JSON Marshal",
		)
	}
	return response, nil
}
