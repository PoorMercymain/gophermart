package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo"

	"github.com/PoorMercymain/gophermart/internal/domain"
	"github.com/PoorMercymain/gophermart/pkg/util"
)

func CheckAuth(ur domain.UserRepository) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if _, ok := c.Request().Header["Authorization"]; ok {
				c.Response().WriteHeader(http.StatusUnauthorized) // authorization with header is forbidden
				return nil
			}

			cookie, err := c.Request().Cookie("jwt")
			if err != nil && !errors.Is(err, http.ErrNoCookie) {
				c.Response().WriteHeader(http.StatusBadRequest)
				return err
			} else if errors.Is(err, http.ErrNoCookie) {
				c.Response().WriteHeader(http.StatusUnauthorized)
				return err
			}

			cookieString := cookie.Value

			user, err := GetUserFromJWT(cookieString)
			if err != nil {
				c.Response().WriteHeader(http.StatusBadRequest)
				return err
			}

			passwordHash, err := ur.GetPasswordHash(c.Request().Context(), user.Login)
			if err != nil {
				c.Response().WriteHeader(http.StatusUnauthorized)
				util.LogInfoln(err)
				return err
			}

			if passwordHash != user.Password {
				c.Response().WriteHeader(http.StatusUnauthorized)
				return errors.New("password isn`t correct")
			}

			ctx := context.WithValue(c.Request().Context(), domain.Key("login"), user.Login)

			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}

func GetUserFromJWT(tokenString string) (domain.User, error) {
	claims := jwt.MapClaims{
		"str": "",
	}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte("ultrasecretkey"), nil
	})
	if err != nil {
		util.LogInfoln("Couldn't parse", err)
		return domain.User{}, err
	}

	if !token.Valid {
		util.LogInfoln("Token isn`t valid")
		return domain.User{}, errors.New("invalid token")
	}

	util.LogInfoln("Token is valid")
	util.LogInfoln(claims["str"])

	userSlice := strings.Split(claims["str"].(string), " ")
	if len(userSlice) < 1 {
		util.LogInfoln("incorrect jwt")
		return domain.User{}, errors.New("incorrect jwt")
	}

	user := domain.User{Login: userSlice[0], Password: userSlice[1]}

	return user, nil
}
