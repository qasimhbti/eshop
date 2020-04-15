package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/eshop/config"
	"github.com/eshop/pkg/httperrors"
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
)

type modifyCartHandler struct {
	db          *mongo.Database
	config      *config.Configs
	redisClient *redis.Client
	cartManager interface {
		calcCartTotAmount(items []*item) *cart
	}
	productManager interface {
		get(db *mongo.Database, itemName string) (*product, error)
	}
}

func (h *modifyCartHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	handleHTTP(w, req, "Modify Cart Handler", h.handle)
}

func (h *modifyCartHandler) handle(req *http.Request) ([]byte, error) {
	claims, err := protectedEndpoint(h.config.JWTAccessSecretKey, req)
	if err != nil {
		return nil, errors.WithMessage(
			httperrors.WithCode(err, http.StatusForbidden),
			"Invalid Login Credentials",
		)
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

	newItems, err := h.getRequestData(req)
	if err != nil {
		return nil, errors.WithMessage(
			httperrors.WithCode(err, http.StatusBadRequest),
			"get request data",
		)
	}

	cart := h.cartManager.calcCartTotAmount(newItems)

	cartDetial, err := json.Marshal(cart)
	if err != nil {
		return nil, errors.WithMessage(
			httperrors.WithCode(err, http.StatusUnprocessableEntity),
			"cart-JSON Marshal",
		)
	}

	log.Println("cart detail :", string(cartDetial))
	err = h.redisClient.Set(claims.UserName, string(cartDetial), time.Hour).Err()
	if err != nil {
		return nil, errors.WithMessage(
			httperrors.WithCode(err, http.StatusInternalServerError),
			"cart-save to redis",
		)
	}
	return cartDetial, nil
}

func (h *modifyCartHandler) getRequestData(req *http.Request) ([]*item, error) {
	items := []*item{}
	err := json.NewDecoder(req.Body).Decode(&items)
	if err != nil {
		if err.Error() == "EOF" {
			return nil, errors.Wrap(err, "cart is empty")
		}
		return nil, errors.Wrap(err, "read HTTP request Body")
	}

	for k, v := range items {
		item := *v
		prd, err := h.productManager.get(h.db, item.ProductCode)
		if err != nil {
			log.Printf("error while getting price for item :%s", item.ProductCode)
		}
		items[k].Price = prd.Price
	}
	return items, nil
}
