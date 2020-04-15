package main

import (
	"encoding/json"
	"net/http"

	"github.com/eshop/config"
	"github.com/eshop/pkg/httperrors"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
)

type productHTTPHandler struct {
	db             *mongo.Database
	config         *config.Configs
	productManager interface {
		getALL(db *mongo.Database) ([]*product, error)
	}
}

func (h *productHTTPHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	handleHTTP(w, req, "Product Handler", h.handle)
}

func (h *productHTTPHandler) handle(req *http.Request) ([]byte, error) {
	_, err := protectedEndpoint(h.config.JWTAccessSecretKey, req)
	if err != nil {
		return nil, errors.WithMessage(
			httperrors.WithCode(err, http.StatusForbidden),
			"Invalid Login Credentials",
		)
	}

	products, err := h.productManager.getALL(h.db)
	if err != nil {
		return nil, errors.WithMessage(err, "get product details")
	}

	response, _ := json.Marshal(&products)
	return response, nil
}
