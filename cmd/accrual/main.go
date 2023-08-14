package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/asaskevich/govalidator"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/PoorMercymain/gophermart/internal/accrual/calculator"
	"github.com/PoorMercymain/gophermart/internal/accrual/config"
	routerAccrual "github.com/PoorMercymain/gophermart/internal/accrual/router"
	"github.com/PoorMercymain/gophermart/internal/accrual/storage"
	"github.com/PoorMercymain/gophermart/pkg/util"
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

	timeoutInterval := 5 * time.Second
	shutdownCtx, cancel := context.WithTimeout(ctx, timeoutInterval)
	defer cancel()

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
	util.GetLogger().Infoln("accrual shutdown signal")

	wg.Wait()

	start := time.Now()

	util.GetLogger().Infoln("before accrual shutdown")
	if err := router.Shutdown(shutdownCtx); err != nil {
		util.GetLogger().Infoln("shutdown:", err)
		return
	} else {
		cancel()
	}

	util.GetLogger().Infoln("accrual after shutdown")
	longShutdown := make(chan struct{}, 1)

	timeoutInterval = 3 * time.Second

	go func() {
		time.Sleep(timeoutInterval)
		longShutdown <- struct{}{}
	}()

	select {
	case <-shutdownCtx.Done():
		util.GetLogger().Infoln("accrual shutdownCtx done:", shutdownCtx.Err().Error())
		util.GetLogger().Infoln(time.Since(start))
		return
	case <-longShutdown:
		util.GetLogger().Infoln("accrual long shutdown finished")
		return
	}

}
