package middleware

import (
	"context"
	"github.com/PoorMercymain/gophermart/pkg/util"
	"github.com/labstack/echo"

	"github.com/PoorMercymain/gophermart/internal/domain"
)

func AddAccrualCommunicatorToCtx(ac domain.Communicator) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			util.GetLogger().Infoln("in accrual communicator")
			ctx := context.WithValue(c.Request().Context(), domain.Key("accrual_communicator"), ac)

			c.SetRequest(c.Request().WithContext(ctx))
			util.GetLogger().Infoln("accrual communicator set to ", ac)
			return next(c)
		}
	}
}
