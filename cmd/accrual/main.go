package main

import (
	"context"
	"github.com/PoorMercymain/gophermart/internal/accrual/calculator"
	"github.com/PoorMercymain/gophermart/internal/accrual/config"
	routerAccrual "github.com/PoorMercymain/gophermart/internal/accrual/router"
	"github.com/PoorMercymain/gophermart/internal/accrual/storage"
	"github.com/PoorMercymain/gophermart/pkg/util"
	"github.com/asaskevich/govalidator"
	_ "github.com/jackc/pgx/v5/stdlib"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {

	util.InitLogger()
	govalidator.SetFieldsRequiredByDefault(true)

	host, dbURI := config.GetAccrualServerConfig()

	dbs, err := storage.NewDBStorage(*dbURI)

	if err != nil {
		util.GetLogger().Infoln(err)
		return
	}

	defer dbs.ClosePool()
	var wg sync.WaitGroup

	ctx := context.Background()

	err = calculator.ProcessUnprocessedOrders(ctx, dbs, &wg)
	if err != nil {
		util.GetLogger().Infoln(err)
		return
	}

	router := routerAccrual.Router(dbs, &wg)
	go router.Start(*host)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	<-sigChan
	util.GetLogger().Infoln("got signal")

	wg.Wait()

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
