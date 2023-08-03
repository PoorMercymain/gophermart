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

	ctx := context.Background()

	err = calculator.ProcessUnprocessedOrders(ctx, dbs)
	if err != nil {
		util.GetLogger().Infoln(err)
		return
	}

	router := routerAccrual.Router(dbs)
	err = router.Start(*host)

	if err != nil {
		util.GetLogger().Infoln(err)
		return
	}

}
