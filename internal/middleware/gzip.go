package middleware

import (
	"compress/gzip"
	"net/http"

	"github.com/PoorMercymain/gophermart/pkg/util"
	"github.com/labstack/echo"
)

func UseGzipReader() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			util.GetLogger().Infoln("in gzip")
			if len(c.Request().Header.Values("Content-Encoding")) == 0 {
				util.GetLogger().Infoln("no gzip")
				return next(c)
			}
			for i, headerValue := range c.Request().Header.Values("Content-Encoding") {
				if headerValue == "gzip" {
					break
				}
				util.LogInfoln(i, (len(c.Request().Header.Values("Content-Type")) - 1))
				if i == (len(c.Request().Header.Values("Content-Type")) - 1) {
					util.GetLogger().Infoln("no gzip")
					return next(c)
				}
			}

			util.LogInfoln("чего")
			gzipReader, err := gzip.NewReader(c.Request().Body)
			if err != nil {
				util.GetLogger().Infoln("no")
				c.Response().WriteHeader(http.StatusBadRequest)
				return err
			}
			c.Request().Body.Close()

			c.Request().Body = gzipReader

			return next(c)
		}
	}
}
