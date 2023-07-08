package repository

import (
	"context"

	"github.com/PoorMercymain/gophermart/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type user struct {
	mongoCollection *mongo.Collection
}

func NewUser(mongoCollection *mongo.Collection) *user {
	return &user{mongoCollection: mongoCollection}
}

func (r *user) Register(ctx context.Context, user domain.User, uniqueLoginErrorChan chan error) error {
	_, err := r.mongoCollection.InsertOne(ctx, user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			uniqueLoginErrorChan <- err
			close(uniqueLoginErrorChan)
		}
	}
	return err
}

func (r *user) GetPasswordHash(ctx context.Context, login string) (string, error) {
	var user domain.User
	filter := bson.M{"login": login}
	err := r.mongoCollection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		return "", err
	}
	return user.Password, nil
}
