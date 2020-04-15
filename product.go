package main

import (
	"context"
	"log"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/mgo.v2/bson"
)

const productCollection = "products"

type product struct {
	ProductCode string  `json:"product_code" bson:"product_code"`
	Name        string  `json:"name" bson:"name"`
	Price       float64 `json:"price" bson:"price"`
}

type productManagerImpl struct{}

func (m *productManagerImpl) getALL(db *mongo.Database) ([]*product, error) {
	var products []*product
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{}
	cursor, err := db.
		Collection(productCollection).
		Find(ctx, filter)
	if err != nil {
		return nil, errors.WithMessage(err, "get all")
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(ctx) {
		var data product
		err := cursor.Decode(&data)
		if err != nil {
			log.Printf("error while decoding : %v", err)
			continue
		}
		products = append(products, &data)
	}
	return products, nil
}

func (m *productManagerImpl) get(db *mongo.Database, itemCode string) (*product, error) {
	var prd *product
	filter := bson.M{"product_code": itemCode}
	err := db.
		Collection(productCollection).
		FindOne(context.Background(), filter).
		Decode(&prd)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("item : %s not available", itemCode)
			return nil, nil
		}
		return nil, errors.WithMessage(err, "get products")
	}
	return prd, nil
}
