package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"net/http"

	"github.com/PoorMercymain/gophermart/internal/handler"
	"github.com/PoorMercymain/gophermart/internal/middleware"
	"github.com/PoorMercymain/gophermart/internal/repository"
	"github.com/PoorMercymain/gophermart/internal/service"
	"github.com/PoorMercymain/gophermart/pkg/util"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/labstack/echo"
	"github.com/pressly/goose/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NewPG(DSN string) (*pgxpool.Pool, error) {
	pg, err := sql.Open("pgx/v5", DSN)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	err = goose.SetDialect("postgres")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	err = pg.PingContext(context.Background())
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	err = goose.Run("up", pg, "./pkg/migrations")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	pg.Close()

	config, err := pgxpool.ParseConfig(DSN)
	if err != nil {
		fmt.Println("Error parsing DSN:", err)
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		fmt.Println("Error creating pgxpool:", err)
		return nil, err
	}

	fmt.Println("норм", pool, err)
	return pool, err
}

func router(pool *pgxpool.Pool) *echo.Echo {
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

	ur := repository.NewUser(collection, pool)
	err = ur.Ping(context.Background())
	if err != nil {
		util.LogInfoln(err)
	} else {
		util.LogInfoln("норм")
	}
	us := service.NewUser(ur)
	uh := handler.NewUser(us)

	e.POST("/api/user/register", uh.Register, middleware.UseGzipReader())
	e.POST("/api/user/login", uh.Authenticate, middleware.UseGzipReader())
	e.POST("/api/user/orders", uh.AddOrder, middleware.UseGzipReader(), middleware.CheckAuth(ur))
	e.GET("/api/user/orders", uh.ReadOrders, middleware.UseGzipReader(), middleware.CheckAuth(ur))
	e.GET("/test", func(c echo.Context) error { return c.String(http.StatusOK, "Hello, World!") }, middleware.UseGzipReader(), middleware.CheckAuth(ur))
	return e
}

func main() {
	dsn := flag.String("d", "", "postgres DSN")
	flag.Parse()

	pool, err := NewPG(*dsn)
	if err != nil {
		util.LogInfoln(err)
		return
	}
	defer pool.Close()
	router(pool).Start(":8080")
}
