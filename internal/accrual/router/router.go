package router

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	eMiddleware "github.com/labstack/echo/v4/middleware"

	"github.com/PoorMercymain/gophermart/internal/accrual/handler"
	"github.com/PoorMercymain/gophermart/internal/accrual/interfaces"
	"github.com/PoorMercymain/gophermart/internal/accrual/middleware"
)

const RequestsPerSecond = 0.5
const RequestsAtSameTime = 10
const RetryAfterInterval = 60

func Router(dbs interfaces.Storage) *echo.Echo {

	e := echo.New()
	sh := handler.NewStorageHandler(dbs)

	rateLimiterConfig := eMiddleware.RateLimiterConfig{
		Skipper: eMiddleware.DefaultSkipper,
		Store: eMiddleware.NewRateLimiterMemoryStoreWithConfig(
			eMiddleware.RateLimiterMemoryStoreConfig{Rate: RequestsPerSecond, Burst: RequestsAtSameTime, ExpiresIn: time.Minute},
		),
		IdentifierExtractor: func(ctx echo.Context) (string, error) {
			id := ctx.RealIP()
			return id, nil
		},
		ErrorHandler: func(context echo.Context, err error) error {
			return context.String(http.StatusForbidden, "Forbidden")
		},
		DenyHandler: func(context echo.Context, identifier string, err error) error {
			context.Response().Header().Set("Retry-After", strconv.Itoa(RetryAfterInterval))
			return context.String(http.StatusTooManyRequests, fmt.Sprintf("No more than %v requests per minute allowed", RequestsPerSecond*60))
		},
	}

	e.Use(eMiddleware.RateLimiterWithConfig(rateLimiterConfig))
	e.Use(middleware.UseGzipReader())

	e.GET("/api/orders/:number", sh.ProcessGetOrdersRequest)
	e.POST("/api/orders", sh.ProcessPostOrdersRequest)
	e.POST("/api/goods", sh.ProcessPostGoodsRequest)

	e.GET("/test", func(c echo.Context) error { return c.String(http.StatusOK, "Test Accrual") })

	return e
}
