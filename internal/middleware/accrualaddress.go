package middleware

import (
	"context"

	"github.com/PoorMercymain/gophermart/internal/domain"
	"github.com/labstack/echo"
)

func AddAccrualAddressToCtx(accrualAddress string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := context.WithValue(c.Request().Context(), domain.Key("accrual_address"), accrualAddress)

			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}
