package router

import (
	"net/http"

	"github.com/labstack/echo"

	"github.com/PoorMercymain/gophermart/internal/accrual/handler"
	"github.com/PoorMercymain/gophermart/internal/accrual/interfaces"
	"github.com/PoorMercymain/gophermart/internal/accrual/middleware"
)

func Router(dbs interfaces.Storage) *echo.Echo {

	e := echo.New()

	sh := handler.NewStorageHandler(dbs)

	e.GET("/api/orders/:number", sh.ProcessGetOrdersRequest, middleware.UseGzipReader())
	e.POST("/api/orders", sh.ProcessPostOrdersRequest, middleware.UseGzipReader())
	e.POST("/api/goods", sh.ProcessPostGoodsRequest, middleware.UseGzipReader())

	e.GET("/test", func(c echo.Context) error { return c.String(http.StatusOK, "Test Accrual") })

	return e
}
