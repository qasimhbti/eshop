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

type createCartHandler struct {
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

func (h *createCartHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	handleHTTP(w, req, "Create Cart Handler", h.handle)
}

func (h *createCartHandler) handle(req *http.Request) ([]byte, error) {
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

	if userCart != "" {
		log.Printf("cart already existed for user: %s", claims.UserName)
		return []byte(userCart), nil
	}

	items, err := h.getRequestData(req)
	if err != nil {
		return nil, errors.WithMessage(
			httperrors.WithCode(err, http.StatusBadRequest),
			"get request data",
		)
	}

	cart := h.cartManager.calcCartTotAmount(items)

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

type item struct {
	ProductCode string  `json:"product_code"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
}

func (h *createCartHandler) getRequestData(req *http.Request) ([]*item, error) {
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
