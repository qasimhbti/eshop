package main

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/mgo.v2/bson"
)

const usersCollection = "users"

type loginCred struct {
	UserName string `json:"username"`
	Password string `json:"password"`
}

type loginCredentialManagerImpl struct{}

func (l *loginCredentialManagerImpl) get(userName string, db *mongo.Database) (*loginCred, error) {
	var cred *loginCred
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"username": userName}
	err := db.
		Collection(usersCollection).
		FindOne(ctx, filter).
		Decode(&cred)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("userName : %s does not found", userName)
		}
		return nil, errors.WithMessage(err, "login credentials")
	}
	return cred, nil
}
