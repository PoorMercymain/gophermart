package main

import (
	"context"

	"github.com/PoorMercymain/gophermart/internal/handler"
	"github.com/PoorMercymain/gophermart/internal/repository"
	"github.com/PoorMercymain/gophermart/internal/service"
	"github.com/PoorMercymain/gophermart/pkg/util"
	"github.com/labstack/echo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func router() *echo.Echo {
	util.InitLogger()

	e := echo.New()

	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	ctx := context.Background()
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		util.LogInfoln(err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		util.LogInfoln(err)
	}

	collection := client.Database("gophermart").Collection("users")

	indexModel := mongo.IndexModel{
		Keys:    bson.M{"login": 1},
		Options: options.Index().SetUnique(true),
	}

	_, err = collection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		util.LogInfoln(err)
	}

	ur := repository.NewUser(collection)
	us := service.NewUser(ur)
	uh := handler.NewUser(us)

	e.POST("/api/user/register", uh.Register)
	e.POST("/api/user/login", uh.Authenticate)
	return e
}

func main() {
	router().Start(":8080")
}
