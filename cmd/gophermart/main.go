package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/PoorMercymain/gophermart/internal/domain"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/PoorMercymain/gophermart/internal/conf"
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
	pg, err := sql.Open("pgx", DSN)
	if err != nil {
		util.GetLogger().Infoln(err)
		return nil, err
	}
	err = goose.SetDialect("postgres")
	if err != nil {
		util.GetLogger().Infoln(err)
		return nil, err
	}

	err = pg.PingContext(context.Background())
	if err != nil {
		util.GetLogger().Infoln(err)
		return nil, err
	}

	const migrationsPath = "./internal/repository/migrations"

	err = goose.Run("up", pg, migrationsPath)
	if err != nil {
		util.GetLogger().Infoln(err)
		return nil, err
	}
	pg.Close()

	config, err := pgxpool.ParseConfig(DSN)
	if err != nil {
		util.GetLogger().Infoln(err)
		fmt.Println("Error parsing DSN:", err)
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		util.GetLogger().Infoln(err)
		fmt.Println("Error creating pgxpool:", err)
		return nil, err
	}

	util.GetLogger().Infoln("норм1", pool, err)
	fmt.Println("норм", pool, err)
	return pool, err
}

func router(pool *pgxpool.Pool, mongoURI string, accrualAddress string, wg *sync.WaitGroup) *echo.Echo {
	e := echo.New()

	clientOptions := options.Client().ApplyURI(mongoURI)
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

	ac := domain.NewAccrualCommunicator(accrualAddress)

	util.GetLogger().Infoln("---------------------")
	err = uh.HandleStartup(accrualAddress, wg)
	if err != nil {
		util.GetLogger().Infoln(err)
	}
	util.GetLogger().Infoln("---------------------")

	e.POST("/api/user/register", uh.Register, middleware.UseGzipReader())
	e.POST("/api/user/login", uh.Authenticate, middleware.UseGzipReader())
	e.POST("/api/user/orders", uh.AddOrder(wg), middleware.UseGzipReader(), middleware.CheckAuth(ur), middleware.AddAccrualCommunicatorToCtx(ac))
	e.GET("/api/user/orders", uh.ReadOrders, middleware.UseGzipReader(), middleware.CheckAuth(ur))
	e.GET("/api/user/balance", uh.ReadBalance, middleware.UseGzipReader(), middleware.CheckAuth(ur))
	e.POST("/api/user/balance/withdraw", uh.AddWithdrawal, middleware.UseGzipReader(), middleware.CheckAuth(ur))
	e.GET("/api/user/withdrawals", uh.ReadWithdrawals, middleware.UseGzipReader(), middleware.CheckAuth(ur))
	e.GET("/test", func(c echo.Context) error { return c.String(http.StatusOK, "Hello, World!") }, middleware.UseGzipReader(), middleware.CheckAuth(ur))
	return e
}

func main() {
	util.InitLogger()

	util.GetLogger().Infoln("логгер запустился")

	config := conf.GetServerConfig()

	util.GetLogger().Infoln(config)

	pool, err := NewPG(config.DatabaseURI)
	if err != nil {
		util.LogInfoln(err)
		return
	}
	defer pool.Close()
	util.GetLogger().Infoln("дошел до router")
	var wg sync.WaitGroup

	r := router(pool, config.MongoURI, config.AccrualAddress, &wg)
	go r.Start(config.ServerAddress)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	<-sigChan
	util.GetLogger().Infoln("got signal")
	wgDone := make(chan struct{}, 1)

	go func() {
		wg.Wait()
		wgDone <- struct{}{}
	}()

	select {
	case <-wgDone:
		util.GetLogger().Infoln("wg were done")
	case <-time.After(time.Second * 5):
		util.GetLogger().Infoln("wg were not done")
	}

	util.GetLogger().Infoln("дальше wg")
	start := time.Now()

	timeoutInterval := 5 * time.Second

	shutdownCtx, cancel := context.WithTimeout(context.Background(), timeoutInterval)
	defer cancel()

	util.GetLogger().Infoln("дошел до shutdown")
	if err := r.Shutdown(shutdownCtx); err != nil {
		util.GetLogger().Infoln("shutdown:", err)
		return
	} else {
		cancel()
	}

	util.GetLogger().Infoln("прошел shutdown")
	longShutdown := make(chan struct{}, 1)

	go func() {
		time.Sleep(3 * time.Second)
		longShutdown <- struct{}{}
	}()

	select {
	case <-shutdownCtx.Done():
		util.GetLogger().Infoln("shutdownCtx done:", shutdownCtx.Err().Error())
		util.GetLogger().Infoln(time.Since(start))
		return
	case <-longShutdown:
		util.GetLogger().Infoln("long shutdown finished")
		return
	}
}
